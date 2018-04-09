package main // import "github.com/danielparks/goldap"

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/go-ldap/ldap"
)

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
	hostname := "ldap.puppetlabs.com"
	port := 389
	userDn := fmt.Sprintf("uid=%s,ou=users,dc=puppetlabs,dc=com", os.Getenv("puppetpass_username"))
	password := os.Getenv("puppetpass_password")
	searchBase := "dc=puppetlabs,dc=com"

	if len(os.Args) < 2 {
		log.Fatalln("usage: goldap <query> [attribute attribute...]")
	}

	query := os.Args[1]
	desiredAttributes := os.Args[2:]

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
	err = conn.Bind(userDn, password)
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
