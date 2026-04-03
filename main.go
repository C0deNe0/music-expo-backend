package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Request struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func main() {
	// create downloads folder if not exists
	os.MkdirAll("downloads", os.ModePerm)

	app := fiber.New()

	app.Post("/extract", func(c *fiber.Ctx) error {
		var req Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		if req.URL == "" {
			return c.Status(400).JSON(fiber.Map{"error": "url required"})
		}

		// basic validation
		if !(contains(req.URL, "watch?v=") || contains(req.URL, "youtu.be/")) {
			return c.Status(400).JSON(fiber.Map{"error": "invalid YouTube video URL"})
		}

		// run in background
		go processAudio(req.URL, req.Name)

		return c.JSON(fiber.Map{
			"status": "processing started",
		})
	})
	app.Static("/", "./static")             // serve the UI folder
	app.Static("/downloads", "./downloads") // serve generated MP3s
	app.Get("/files", func(c *fiber.Ctx) error {
		files, err := os.ReadDir("downloads")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		mp3s := []string{}
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".mp3") {
				mp3s = append(mp3s, f.Name())
			}
		}

		return c.JSON(mp3s)
	})

	log.Fatal(app.Listen(":8080"))
}

// contains helper
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > len(sub) && (string(s[:len(sub)]) == sub || contains(s[1:], sub)))
}

func processAudio(url, name string) {
	fmt.Println("Processing:", url)

	output := fmt.Sprintf("downloads/%s.%%(ext)s", name)

	// full path to standalone yt-dlp binary
	cmd := exec.Command(
		"/usr/local/bin/yt-dlp",
		"-x",                    // extract audio
		"--audio-format", "mp3", // convert to mp3
		"--audio-quality", "0", // best quality
		"-o", output,
		url,
	)
	// capture stdout and stderr
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("STDOUT:", out.String())
		fmt.Println("STDERR:", stderr.String())
		return
	}

	fmt.Println("Done:", url)
}
