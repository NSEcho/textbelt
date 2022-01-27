# textbelt

Golang library for textbelt

# Usage

```golang
package main

import (
	"fmt"
	"time"

	"github.com/lateralusd/textbelt"
)

func main() {
	texter := textbelt.New(
		textbelt.WithKey("textbelt"),
		textbelt.WithTimeout(3*time.Second),
	)

	rem, err := texter.Quota()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Remaining messages: %d\n", rem)

	msg, err := texter.Send("+5555555555", "test message")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Message id is %s\n", msg)

	status, err := texter.Status(msg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Message \"%s\" status is \"%s\"\n", msg, status)

	otp, err := texter.GenerateOTP("+5555555555", "testuserid")
	if err != nil {
		panic(err)
	}

	valid, err := texter.VerifyOTP(otp, "testuserid")
	if err != nil {
		panic(err)
	}

	fmt.Println("OTP is", valid)
}
```
