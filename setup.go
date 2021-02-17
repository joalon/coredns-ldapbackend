package ldapbackend

import (
    "fmt"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

    ldap "github.com/go-ldap/ldap/v3"
)

func init() { plugin.Register("ldapbackend", setup) }

func setup(c *caddy.Controller) error {
	c.Next()
    c.NextArg()
    zone := c.Val()

    c.NextArg()
    ldapURL := c.Val()

    var (
        binddn string
        password string
        basedn string
    )
    for c.NextBlock() {
        arg := c.Val()

        switch arg {
        case "binddn":
            c.NextArg()
            binddn = c.Val()
            c.NextArg()
            password = c.Val()
        case "basedn":
            c.NextArg()
            basedn = c.Val()
        default:
            return plugin.Error("ldapbackend", c.Errf("unexpected argument: %v", arg))
        }
    }

    //l, err := ldap.DialURL(ldapURL, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
    l, err := ldap.DialURL(ldapURL)
    if err != nil {
        return plugin.Error("ldapbackend", err)
    }

    err = l.Bind(binddn, password)
    if err != nil {
        return plugin.Error("ldapbackend", err)
    }

    fmt.Println(zone)
    fmt.Println("setup done!")
    ldapbackend := LdapBackend{LdapConn: l, BaseDn: basedn}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
        ldapbackend.Next = next
		return ldapbackend
	})

	return nil
}
