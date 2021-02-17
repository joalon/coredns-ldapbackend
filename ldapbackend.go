package ldapbackend

import (
	"context"
	"net"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	ldap "github.com/go-ldap/ldap/v3"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("ldapbackend")

type LdapBackend struct {
    Next        plugin.Handler
    LdapConn    *ldap.Conn
    BaseDn  string
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


    var (
        ldapRecordPrefix string
    )

    switch r.Question[0].Qtype {
    case 1:
        log.Debug("Got A")
        ldapRecordPrefix = "aRecord="
    case 12:
        log.Debug("Got PTR")
        ldapRecordPrefix = "pTRRecord="
    }

    ldapSearchFilter := "(&(objectClass=dNSZone))"

    searchRequest := ldap.NewSearchRequest(
        l.BaseDn,
        ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
        ldapSearchFilter,
        []string{"dn", "cn"},
        nil,
    )

    sr, err := l.LdapConn.Search(searchRequest)
    if err != nil {
            log.Fatal(err)
    }

    if len(sr.Entries) == 0 {
        return plugin.NextOrFailure(l.Name(), l.Next, ctx, w, r)
    }

    state := request.Request{W: w, Req: r}

    for _, entry := range sr.Entries {
        //log.Debug(entry.DN, entry.GetAttributeValue("cn"))
        log.Debug(entry.DN)

        if strings.HasPrefix(entry.DN, ldapRecordPrefix) {
            // Found the record! Just trim it down
            tmpStrWithSuffix := strings.TrimPrefix(entry.DN, ldapRecordPrefix)
            finalIp := strings.TrimSuffix(tmpStrWithSuffix, "," + l.BaseDn)

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

