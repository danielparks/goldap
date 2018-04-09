package main // import "github.com/danielparks/goldap"

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/go-ldap/ldap"
	getopt "github.com/pborman/getopt/v2"
)

var (
	hostname     = "ldap.puppetlabs.com"
	port         = 389
	searchBase   = "dc=puppetlabs,dc=com"
	userUID      = os.Getenv("puppetpass_username")
	userDN       = fmt.Sprintf("uid=%s,ou=users,dc=puppetlabs,dc=com", userUID)
	userPassword = os.Getenv("puppetpass_password")
)

func init() {
	getopt.FlagLong(&hostname, "hostname", 'h', "hostname of the LDAP server")
	getopt.FlagLong(&port, "port", 0, "port of LDAP server")
	getopt.FlagLong(&searchBase, "search-base", 'b', "base fo LDAP search")
	getopt.FlagLong(&userDN, "user-dn", 'u', "DN of user to connect as")
	getopt.FlagLong(&userPassword, "password", 'p', "Password of user to connect as")
}

func printEntry(entry *ldap.Entry) {
	fmt.Printf("dn: %s\n", entry.DN)
	for _, attribute := range entry.Attributes {
		for _, value := range attribute.Values {
			if len(value) > 100 {
				fmt.Printf("%s: %s\n", attribute.Name, "<long>")
			} else {
				fmt.Printf("%s: %s\n", attribute.Name, value)
			}
		}
	}
}

func main() {
	getopt.Parse()
	args := getopt.Args()

	if len(args) < 1 {
		log.Fatalln("usage: goldap <query> [attribute attribute...]")
	}

	query := args[0]
	desiredAttributes := args[1:]

	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Upgrade to TLS
	tlsConfig := &tls.Config{
		ServerName:         hostname,
		InsecureSkipVerify: false,
	}
	err = conn.StartTLS(tlsConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Authenticate
	err = conn.Bind(userDN, userPassword)
	if err != nil {
		log.Fatal(err)
	}

	// Search
	searchRequest := ldap.NewSearchRequest(
		searchBase, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		query,
		desiredAttributes,
		nil,
	)

	response, err := conn.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}

	if len(response.Entries) > 0 {
		printEntry(response.Entries[0])

		if len(response.Entries) > 1 {
			for _, entry := range response.Entries[1:] {
				fmt.Print("\n")
				printEntry(entry)
			}
		}
	}
}
