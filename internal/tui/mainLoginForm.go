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
	"github.com/haukened/tsky/internal/debug"
	"github.com/haukened/tsky/internal/messages"
	"github.com/haukened/tsky/internal/utils"
)

var (
	//lint:ignore ST1005 I want it that way
	ErrHandleDoesNotResolve = errors.New("Handle does not resolve")
	//lint:ignore ST1005 I want it that way
	ErrInvalidHandle = errors.New("Invalid handle")
	//lint:ignore ST1005 I want it that way
	ErrPasswordEmpty = errors.New("Password cannot be empty")
	//lint:ignore ST1005 I want it that way
	ErrInvalidPassword = errors.New("Invalid password, please use an app password not your primary account password")
	//lint:ignore ST1005 I want it that way
	ErrEmailDomainNotExist = errors.New("Email domain does not exist")
	//lint:ignore ST1005 I want it that way
	ErrDisallowedTLD = errors.New("Disallowed TLD")
	// but this one is somehow fine?
	ErrHttpClient = errors.New("HTTP client error")
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
	show bool
}

func initialForm(c *config.Config) *huh.Form {
	return huh.NewForm(
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
}

func NewLoginModel(c *config.Config) NamedModel {
	f := initialForm(c)
	return LoginModel{
		form: f,
		conf: c,
		show: false,
	}
}

func (m LoginModel) Name() string {
	return "login"
}

func (m LoginModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m LoginModel) Update(msg tea.Msg) (NamedModel, tea.Cmd) {
	var cmds []tea.Cmd
	if !authNeeded(m.conf) {
		debug.Debugf("No auth needed, skipping login")
		cmds = append(cmds, messages.Next)
		cmds = append(cmds, messages.StartAuth)
		return m, tea.Batch(cmds...)
	} else {
		m.show = true
	}
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
			cmd = messages.SendErrorMsg(m.form.Errors()[0].Error())
			cmds = append(cmds, cmd)
		} else {
			cmd = messages.ClearStatusMsg()
			cmds = append(cmds, cmd)
		}
	} else {
		// form is completed
		m.conf.Save()
		cmds = append(cmds, messages.Next)
		cmds = append(cmds, messages.StartAuth)
	}
	// return the updated model and the batched commands
	return m, tea.Batch(cmds...)
}

func (m LoginModel) View() string {
	if m.show {
		return m.form.View()
	}
	return ""
}

func authNeeded(c *config.Config) bool {
	if c.Identifier == "" {
		// if we don't have a username we need to auth
		return true
	}
	if c.RefreshJwt == "" {
		// if we don't have a refresh token we need to auth
		return true
	} else {
		// if the refresh token is expired we need to auth
		if utils.IsJwtExpired(c.RefreshJwt) {
			return true
		}
	}
	// if we have both a username and a valid refresh token we don't need to auth
	return false
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
