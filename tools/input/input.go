package input

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func ReadInput(prompt string) (rv string, err error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(prompt)

	rv, err = reader.ReadString('\n')
	if err != nil {
		return
	}
	rv = strings.Replace(rv, "\n", "", -1)
	return
}

func MustReadInput(prompt string) (rv string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(prompt)

	rv, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	rv = strings.Replace(rv, "\n", "", -1)
	return
}
