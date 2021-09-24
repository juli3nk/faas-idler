package main

type Config struct {
	Interval uint32 `default:"30"`
}

type Gateway struct {
	URL          string `default:"http://gateway:8080"`
	UsernameFile string `default:"/run/secrets/basic-auth-user"`
	PasswordFile string `default:"/run/secrets/basic-auth-password"`
}

type Metric struct {
	Host               string `default:"prometheus"`
	Port               int    `default:"9090"`
	InactivityDuration uint32 `split_words:"true" default:"15"`
}
