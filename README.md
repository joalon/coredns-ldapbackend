# CoreDNS LDAP Plugin
A plugin for storing DNS zones in an LDAP backend.

## Compile


## Example

```
.:10053 {
    ldapbackend example.com ldap://ds.example.com {
        baseDn
        bindName
        create-if-not-exists <true>
        tlsConfig
    }
}
```

