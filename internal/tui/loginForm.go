package tui

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/haukened/tsky/internal/config"
	"github.com/haukened/tsky/internal/messages"
	"github.com/haukened/tsky/internal/utils"
)

var (
	ErrHandleDoesNotResolve = errors.New("handle does not resolve")
	ErrInvalidHandle        = errors.New("invalid handle")
	ErrPasswordEmpty        = errors.New("password cannot be empty")
	ErrInvalidPassword      = errors.New("invalid password, please use an app password not your primary account password")
	ErrEmailDomainNotExist  = errors.New("email domain does not exist")
	ErrDisallowedTLD        = errors.New("disallowed TLD")
	ErrHttpClient           = errors.New("http client error")
)

var disallowedTLDs = []string{
	".alt",
	".arpa",
	".example",
	".internal",
	".invalid",
	".local",
	".localhost",
	".onion",
	".test",
}

type LoginModel struct {
	form *huh.Form
	conf *config.Config
	done bool
}

func NewFormModel(c *config.Config) tea.Model {
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Username").
				Value(&c.Identifier).
				Validate(validateIdentifier),
			huh.NewInput().
				EchoMode(huh.EchoModePassword).
				Title("Password").
				Value(&c.AppPassword).
				Validate(validatePassword),
		),
	).WithShowHelp(false).WithShowErrors(false)

	return LoginModel{
		form: f,
		conf: c,
	}
}

func (m LoginModel) Init() tea.Cmd {
	// determine if we even need to show the login form
	if m.conf.Identifier != "" {
		// we have a username
		if m.conf.RefreshJwt != "" {
			// we have a token so we need to check if its valid
			if !utils.IsJwtExpired(m.conf.RefreshJwt) {
				return messages.LoginFinished
			} else {
				// we have a token but its expired
				// so delete it
				m.conf.RefreshJwt = ""
				m.conf.Save()
			}
		}
	}
	return m.form.Init()
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, nil
	}
	var cmds []tea.Cmd
	if m.form.State != huh.StateCompleted {
		// pass the message to the form
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}
		cmds = append(cmds, cmd)
		// get the form help message
		cmd = messages.SendHelpText(m.form.Help().ShortHelpView(m.form.KeyBinds()))
		cmds = append(cmds, cmd)
		// get the form errors
		if len(m.form.Errors()) > 0 {
			cmd = messages.SendStatusErr(m.form.Errors()[0].Error())
			cmds = append(cmds, cmd)
		}
	} else {
		// form is completed
		m.conf.Save()
		m.done = true
		cmds = append(cmds, messages.LoginFinished)
	}
	// return the updated model and the batched commands
	return m, tea.Batch(cmds...)
}

func (m LoginModel) View() string {
	return m.form.View()
}

func validateIdentifier(s string) error {
	if isEmail(s) {
		err := validateEmail(s)
		if err != nil {
			return err
		}
		return nil
	}
	err := validateHandle(s)
	if err != nil {
		return err
	}
	return nil
}

func validatePassword(s string) error {
	if len(s) == 0 {
		return ErrPasswordEmpty
	}
	// It is a best practice for most clients and apps to include a reminder to use an app password
	// when logging in. App passwords usually have the form xxxx-xxxx-xxxx-xxxx, and clients can
	// check against this format to prevent accidental logins with primary passwords
	// (unless the primary password itself has this format).
	re := regexp.MustCompile(`^[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}$`)
	if re.MatchString(s) {
		return nil
	}
	return ErrInvalidPassword
}

func isEmail(s string) bool {
	// make sure it's a valid email address using regex
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return re.MatchString(s)
}

func validateEmail(s string) error {
	// make sure its a real email with an MX record
	if !hasValidMXRecord(s) {
		return ErrEmailDomainNotExist
	}
	return nil
}

func validateHandle(s string) error {
	// https://atproto.com/specs/handle#handle-identifier-syntax
	// A reference regular expression (regex) for the handle syntax is:
	// /^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$/
	re := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

	if !re.MatchString(s) {
		return ErrInvalidHandle
	}

	// check for disallowed TLDs
	// https://atproto.com/specs/handle#additional-non-syntax-restrictions
	for _, tld := range disallowedTLDs {
		if strings.HasSuffix(s, tld) {
			return ErrDisallowedTLD
		}
	}

	// resolve the handle
	err := resolveHandle(s)
	if err != nil {
		return err
	}

	return nil
}

func hasValidMXRecord(email string) bool {
	// split the address at the @ symbol to get the domain
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	domain := parts[1]
	// check if the domain has an MX record
	mxRecords, err := net.LookupMX(domain)
	if err != nil || len(mxRecords) == 0 {
		return false
	}
	return true
}

func resolveHandle(handle string) error {
	// https://atproto.com/specs/handle#handle-resolution
	dnsLocation := fmt.Sprintf("_atproto.%s", handle)
	// dont check for errors on purpose, because if this fails we can still check the HTTPS location
	record, _ := net.LookupTXT(dnsLocation)
	// if we get responses check them
	if len(record) > 0 {
		for _, r := range record {
			// there could be multiple per the DNS spec
			if strings.HasPrefix(r, "did=") {
				// we found a DID
				return nil
			}
		}
	}
	// then check the HTTPS /.well-known/atproto-did file
	httpsLocation := fmt.Sprintf("https://%s/.well-known/atproto-did", handle)
	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, httpsLocation, nil)
	if err != nil {
		return ErrHttpClient
	}
	req.Header.Set("User-Agent", utils.UserAgent())
	resp, err := client.Do(req)
	if err != nil {
		return ErrHandleDoesNotResolve
	}
	defer resp.Body.Close()
	// check the response code
	if resp.StatusCode != http.StatusOK {
		return ErrHandleDoesNotResolve
	}
	// check the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrHandleDoesNotResolve
	}
	// make sure it contains a DID
	if !strings.HasPrefix(string(body), "did=") {
		return ErrHandleDoesNotResolve
	}
	return nil
}
