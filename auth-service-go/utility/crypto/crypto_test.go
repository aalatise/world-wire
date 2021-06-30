package crypto

import (
	"fmt"
	"log"
	"testing"
)

func TestHashPassword(test *testing.T) {
	fmt.Println("Testing")
	hash, err := HashPassword("password")
	if err != nil {
		log.Println(err)
	}
	//fmt.Fprintln(os.Stdout, hash)
	log.Println(hash)

	b := CheckPasswordHash("password", "$2a$14$G/azmKLLtziOZPN5HzKSWeJT3Bdj.DNGXZSYcTvE.WeddVzAsfDbK")
	log.Println(b)
}
