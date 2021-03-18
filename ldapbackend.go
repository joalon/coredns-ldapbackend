package ldapbackend

import (
	"net"
	"context"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"

	clog "github.com/coredns/coredns/plugin/pkg/log"
	ldap "github.com/go-ldap/ldap/v3"
)

var log = clog.NewWithPlugin("ldapbackend")

type LdapBackend struct {
    Next        plugin.Handler
    LdapConn    *ldap.Conn
    BaseDn      string
}

func (l LdapBackend) Name() string {
    return "LdapBackend"
}

func (l LdapBackend) OnShutdown() {
    log.Debug("Closing ldap connection...\n")
    l.LdapConn.Close()
}

func (l LdapBackend) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

    // Validate


    // Search
    searchRequest := ldap.NewSearchRequest(
        l.BaseDn,
        ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
        "(&(objectClass=dNSZone))",
        []string{"dn", "cn", "dNSTTL", "relativeDomainName"},
        nil,
    )

    sr, err := l.LdapConn.Search(searchRequest)
    if err != nil && !ldap.IsErrorAnyOf(err, ldap.LDAPResultNoSuchObject) {
        log.Fatal(err)
    }

    // If not found: try other plugins
    if len(sr.Entries) == 0 {
        log.Debug("Executed search and didn't find any entries")
        return plugin.NextOrFailure(l.Name(), l.Next, ctx, w, r)
    }

    // If found: return result
    state := request.Request{W: w, Req: r}
    for _, entry := range sr.Entries {
        var ldapRecordPrefix string
        switch r.Question[0].Qtype {
        case dns.TypeA:
            log.Debug("Got A")
            ldapRecordPrefix = "aRecord="
        case dns.TypePTR:
            log.Debug("Got PTR")
            ldapRecordPrefix = "pTRRecord="
        }

        if strings.HasPrefix(entry.DN, ldapRecordPrefix) {
            // Found the record! Just trim it down
            tmpStrWithSuffix := strings.TrimPrefix(entry.DN, ldapRecordPrefix)
            finalIp := strings.TrimSuffix(tmpStrWithSuffix, "," + l.BaseDn)

            log.Debug(entry)
            for _, a := range entry.Attributes {
                log.Debug(a)
            }

            header := dns.RR_Header{ Name: state.QName(), Rrtype: dns.TypeA, Class: state.QClass(), Ttl: 86400}

            message := new(dns.Msg)
            message.SetReply(r)
            message.Authoritative = true
            message.Answer = []dns.RR{ &dns.A{ Hdr: header, A: net.ParseIP(finalIp)} }

            log.Debug("Sending: ")
            log.Debug(message)

            w.WriteMsg(message)
            return dns.RcodeSuccess, nil
        }
    }

    return plugin.NextOrFailure(l.Name(), l.Next, ctx, w, r)
}

