package ldapbackend

import (
    "context"

    "github.com/miekg/dns"
	"github.com/coredns/coredns/plugin"
    clog "github.com/coredns/coredns/plugin/pkg/log"
    ldap "github.com/go-ldap/ldap/v3"
)

var log = clog.NewWithPlugin("ldapbackend")

type LdapBackend struct {
    Next        plugin.Handler
    LdapConn    *ldap.Conn
}

func (l LdapBackend) Name() string {
    return "LdapBackend"
}

func (l LdapBackend) OnShutdown() {
    log.Debug("Closing connection...\n")
    l.LdapConn.Close()
}

func (l LdapBackend) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
    log.Debug("Got a question!")
    log.Debug(r)

    // Guess the ldap magic will happen here!!
    searchRequest := ldap.NewSearchRequest(
        "dc=example,dc=org",
        ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
        "(&(objectClass=*))",
        []string{"dn", "cn"},
        nil,
    )

    sr, err := l.LdapConn.Search(searchRequest)
    if err != nil {
            log.Fatal(err)
    }

    for _, entry := range sr.Entries {
        log.Debug("%s: %v\n", entry.DN, entry.GetAttributeValue("cn"))
    }

    return plugin.NextOrFailure(l.Name(), l.Next, ctx, w, r)
}

