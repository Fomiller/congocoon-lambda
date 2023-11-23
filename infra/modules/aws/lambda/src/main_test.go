package main

import (
  "testing"
  alt "github.com/yogeshlonkar/aws-lambda-go-test/local"
)
response, err := alt.Run(alt.Input{
  "Name:" "forrest"
})
if response != expected {
	t.Error()
}

func TestAbc(t *testing.T) {
	t.Error()
}
