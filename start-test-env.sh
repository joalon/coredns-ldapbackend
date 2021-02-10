#!/bin/bash -e

# Entire script taken from the osixia/phpldapadmin repo

docker run --name ldap-service --hostname ldap-service --detach -P osixia/openldap:1.4.0
docker run --name phpldapadmin-service --hostname phpldapadmin-service --link ldap-service:ldap-host --env PHPLDAPADMIN_LDAP_HOSTS=ldap-host --detach osixia/phpldapadmin:0.9.0

OPENLDAP_IP=$(docker inspect -f "{{ .NetworkSettings.IPAddress }}" ldap-service)
PHPLDAP_IP=$(docker inspect -f "{{ .NetworkSettings.IPAddress }}" phpldapadmin-service)

echo "Go to: https://$PHPLDAP_IP"
echo "Login DN: cn=admin,dc=example,dc=org"
echo "Password: admin"
echo "Access the openldap on $OPENLDAP_IP"


