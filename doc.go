// Copyright 2016 The go-overpass AUTHORS. All rights reserved.
//
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package overpass provides a client for using the Overpass API.

Usage:

	import "github.com/serjvanilla/go-overpass"

Construct a new client, then use Query method on the client to
receive result for your OverpassQL queries.

	client := overpass.New()

	//Retrieve relation with all its members, recursively.
	result, _ := client.Query("[out:json];relation(1673881);>>;out body;")
	//Take a note that you should use "[out:json]" in your query for correct work.

Default client uses overpass-api.de endpoint, but you can choose another with
NewWithSettings method.

	client := overpass.NewWithSettings("http://api.openstreetmap.fr/oapi/interpreter/", 1, http.DefaultClient)

You also can use default client directly by calling Query independently.

	result, _ := overpass.Query("[out:json];relation(1673881);>>;out body;")

# Rate limiting

Library respects servers rate limits and will not perform more than one request simultaneously with default client.
With custom client you are able to adjust that value.
*/
package overpass
