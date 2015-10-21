package router

import (
	"testing"
)

func TestContains(t *testing.T) {
	test := Test{t}

	blacklist, err := CreateBlacklistFromFile("blacklistSample.lst")
	test.AssertNil("Failed to create blacklist", err)

	test.AssertTrue("foo@example.net should be on blacklist", blacklist.Contains(`{"email": "foo@example.net", "firstName": "Egon", "lastName": "Björk", "additional": "Test", "language": "de"}`))
	test.AssertTrue("BAR@example.net should be on blacklist", blacklist.Contains(`{"email": "BAR@example.net", "firstName": "Test", "lastName": "Fun", "language": "en"}`))
	test.AssertFalse("stefm@example.com should not be on blacklist", blacklist.Contains(`{"email": "stefm@example.com", "firstName": "Stefan", "lastName": "Martin", "language": "en"}`))
	test.AssertFalse("empty lines should not be in blacklist", blacklist.Contains(`{"email": ""}`))

	test.AssertTrue("mail0@example.org should be on blacklist", blacklist.Contains(`{"email": "mail0@example.org", "firstName": "Stefan", "lastName": "Martin", "language": "en"}`))
	test.AssertTrue("mail1@example.org should be on blacklist", blacklist.Contains(`{"email": "mail1@example.org", "firstName": "Stefan", "lastName": "Martin", "language": "en"}`))
	test.AssertTrue("mail2@example.org should be on blacklist", blacklist.Contains(`{"email": "mail2@example.org", "firstName": "Stefan", "lastName": "Martin", "language": "en"}`))
	test.AssertTrue("mail3@example.org should be on blacklist", blacklist.Contains(`{"email": "mail3@example.org", "firstName": "Stefan", "lastName": "Martin", "language": "en"}`))
	test.AssertFalse("mail4@example.org should not be on blacklist", blacklist.Contains(`{"email": "mail4@example.org", "firstName": "Stefan", "lastName": "Martin", "language": "en"}`))
}
