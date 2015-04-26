// This example demonstrates using the wrapio to encrypt and decrypt
// io.Readers.
package wrapio_test

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"github.com/icub3d/gop/wrapio"
)

func Example_encryptDecrypt() {
	// These are the messages we'll encrypt and decrypt.
	messages := []string{
		"this is the first message.",
		"I wonder about this encryption thing. Is it safe?",
		"01234567890123456789012345678901",
		"012345678901234567890123456789012",
		"0123456789012345678901234567890",
		"See what I did there?",
	}
	// Make a key.
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Fatalln("failed to generate key:", err)
	}
	for _, message := range messages {
		// We are going to use the AES cipher.
		fmt.Println("original text:", message)
		b, err := aes.NewCipher(key)
		if err != nil {
			log.Fatalln("failed to make aes cipher:", err)
		}
		// CBC needs an initialization vector.
		iv := make([]byte, b.BlockSize())
		if _, err := rand.Read(iv); err != nil {
			log.Fatalln("failed to generate iv:", err)
		}
		bme := cipher.NewCBCEncrypter(b, iv)
		var in, block, last, crypt io.Reader
		// Wrap our message as a reader. In normal circumstances, this
		// may come from a file, a network socket, etc.
		in = strings.NewReader(message)
		// We want to only send slices whose lengths are a multiple of the
		// block size (it will panic otherwise).
		block = wrapio.NewBlockReader(bme.BlockSize(), in)
		// The last block may not have a full block, so we should pad it.
		last = wrapio.NewLastFuncReader(pad, block)
		// Finally, we encrypt the data.
		crypt = wrapio.NewFuncReader(func(p []byte) {
			bme.CryptBlocks(p, p)
		}, last)
		// We'll just use ioutil.ReadAll to get the cipher text. Under
		// normal circumstances, this may be passed along to something
		// that reads the encrypted data. For example, you may pass it to
		// net/http's client.Post() to send the file to a web server.
		ct, err := ioutil.ReadAll(crypt)
		if err != nil {
			log.Fatalln("failed to read all:", err)
		}
		fmt.Println("cipher text:  ", ct)
		// On the other end we'll decrypt it. We'll just use a
		// bytes.Reader but we could wrap net/http's response.Body if we
		// were pulling our encrypted file from a web server.
		bmd := cipher.NewCBCDecrypter(b, iv)
		in = bytes.NewReader(ct)
		// We want to only send slices whose lengths are a multiple of the
		// block size (it will panic otherwise).
		block = wrapio.NewBlockReader(bmd.BlockSize(), in)
		// Decrypt the data we get.
		crypt = wrapio.NewFuncReader(func(p []byte) {
			bmd.CryptBlocks(p, p)
		}, block)
		// The last block may have padding at the end, so remove it.
		last = wrapio.NewLastFuncReader(unpad, crypt)
		// We'll use ReadAll to get the plain text. If we wanted to save
		// it to a file, we could use io.Copy.
		pt, err := ioutil.ReadAll(last)
		if err != nil {
			log.Fatalln("failed to read all:", err)
		}
		fmt.Println("plain text:   ", string(pt))
		fmt.Println()
	}
}

func unpad(p []byte) []byte {
	l := len(p)
	nl := l - 1
	for ; nl >= 0; nl-- {
		if p[nl] == 0x80 {
			break
		}
		if p[nl] != 0x00 || (l-nl) > aes.BlockSize {
			return p
		}
	}
	return p[:nl]
}

func pad(m []byte) []byte {
	l := len(m)
	pl := aes.BlockSize - l%aes.BlockSize
	m = append(m, 0x80)
	for x := 0; x < pl-1; x++ {
		m = append(m, 0x00)
	}
	return m
}
