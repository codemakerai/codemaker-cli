// Copyright 2023 CodeMaker AI Inc. All rights reserved.

package cli

import (
	"flag"
	"fmt"
	"github.com/codemakerai/codemaker-sdk-go/client"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"time"
)

const (
	initialRetryDelay  = 2 * time.Second
	maxRetryDelay      = 60 * time.Second
	processTimeout     = 10 * time.Minute
	nonExponentRetries = 4
	maxExponentRetries = 16
)

type Cli struct {
	logger *zap.SugaredLogger
}

func NewCli() Cli {
	logger := createLogger()
	return Cli{
		logger: logger.Sugar(),
	}
}

func (c *Cli) Run() {
	defer c.logger.Sync()
	c.parseArgs()
}

func (c *Cli) parseArgs() {
	if len(os.Args) < 2 {
		c.printHelp()
	}

	switch os.Args[1] {
	case "generate":
		c.parseGenerateArgs()
		break
	case "configure":
		c.configure()
	case "version":
		c.printVersion()
		break
	default:
		fmt.Printf("Unknown command %s\n", os.Args[1])
		c.printHelp()
	}
}

func (c *Cli) parseGenerateArgs() {
	if len(os.Args) < 3 {
		fmt.Printf("No command specified")
		c.printGenerateHelp()
	}

	switch os.Args[2] {
	case "docs":
		generateDocsCmd := flag.NewFlagSet("generateDocs", flag.ExitOnError)
		lang := generateDocsCmd.String("language", "", "Programming language: Java, Scala, Kotlin")

		err := generateDocsCmd.Parse(os.Args[3:])
		if err != nil {
			c.logger.Error("Could not parse args %v", err)
			os.Exit(1)
		}

		if len(generateDocsCmd.Args()) == 0 {
			c.logger.Errorf("Expected file input")
			fmt.Printf("Usage: codemaker generate docs <file>\n")
			os.Exit(1)
		}

		config, err := createConfig()
		if err != nil {
			c.logger.Errorf("No valid api key found: %v", err)
			os.Exit(1)
		}

		cl := c.createClient(*config)
		input := generateDocsCmd.Args()[0]

		if err := c.generateDocumentation(cl, lang, input); err != nil {
			c.logger.Errorf("Could not generate the documentation %v", err)
		}
		break
	default:
		fmt.Printf("Unknown command %s\n", os.Args[2])
		c.printGenerateHelp()
	}
}

func (c *Cli) generateDocumentation(cl client.Client, lang *string, input string) error {
	return c.walkPath(input, func(file string) error {
		if lang == nil || len(*lang) == 0 {
			actLang, err := LanguageFromExtension(filepath.Ext(file))
			if err != nil {
				return err
			}
			lang = &actLang
		}

		c.logger.Infof("Generating documentation in file %s", file)
		source, err := c.readFile(file)
		if err != nil {
			return err
		}

		output, err := c.process(cl, client.ModeDocument, *lang, source)
		if err != nil {
			return err
		}

		if err := c.writeFile(file, *output); err != nil {
			return err
		}

		return nil
	})
}

func (c *Cli) process(cl client.Client, mode string, lang string, source string) (*string, error) {
	process, err := cl.CreateProcess(&client.CreateProcessRequest{
		Process: client.Process{
			Mode:     mode,
			Language: lang,
			Input: client.Input{
				Source: source,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	retry := 0
	timeout := time.After(processTimeout)
	for {
		status, err := cl.GetProcessStatus(&client.GetProcessStatusRequest{
			Id: process.Id,
		})
		if err != nil {
			return nil, err
		}

		if c.isCompleted(status) {
			break
		}

		select {
		case <-timeout:
			return nil, fmt.Errorf("the task processing had timed out")
		default:
			c.sleep(retry)
			retry++
		}
	}

	output, err := cl.GetProcessOutput(&client.GetProcessOutputRequest{
		Id: process.Id,
	})
	if err != nil {
		return nil, err
	}

	return &output.Output.Source, nil
}

func (c *Cli) sleep(retry int) {
	retry -= nonExponentRetries
	if retry < 0 {
		retry = 0
	}

	if retry > maxExponentRetries {
		retry = maxExponentRetries
	}

	retryDelay := initialRetryDelay * (1 << retry)
	if retryDelay > maxRetryDelay {
		retryDelay = maxRetryDelay
	}

	time.Sleep(retryDelay)
}

func (c *Cli) isCompleted(status *client.GetProcessStatusResponse) bool {
	return status.Status == client.StatusCompleted
}

func (c *Cli) configure() error {
	c.logger.Infof("Configure CLI")

	var apiKey string

	fmt.Print("Enter API Key: ")
	fmt.Scanln(&apiKey)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		c.logger.Errorf("Failed to resolved user home directory %v", err)
		return err
	}

	dir := filepath.Join(homeDir, ".codemaker")
	if err = os.MkdirAll(dir, 0755); err != nil {
		c.logger.Errorf("Failed to create codemaker home directory %v", err)
		return err
	}

	env := map[string]string{apiKeyEnvironmentVariable: apiKey}
	if err = godotenv.Write(env, filepath.Join(dir, "config")); err != nil {
		c.logger.Errorf("Failed to write codemaker configuration %v", err)
		return err
	}
	return nil
}

func (c *Cli) printVersion() {
	c.logger.Infof("CodeMaker CLI version %s (Build %s)", Version, Build)
}

func (c *Cli) printHelp() {
	fmt.Printf("Usage: codemaker <command>\n")
	fmt.Printf("\n")
	fmt.Printf("Commands:\n")
	fmt.Printf(" * generate\n")
	fmt.Printf(" * version\n")
	os.Exit(1)
}

func (c *Cli) printGenerateHelp() {
	fmt.Printf("Usage: codemaker generate <command>\n")
	fmt.Printf("\n")
	fmt.Printf("Commands:\n")
	fmt.Printf(" * docs\n")
	os.Exit(1)
}

func (c *Cli) walkPath(pattern string, visitor func(file string) error) error {
	files, err := c.matchFiles(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := visitor(file); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cli) matchFiles(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func (c *Cli) readFile(file string) (string, error) {
	data, err := os.ReadFile(file)
	return string(data), err
}

func (c *Cli) writeFile(file string, source string) error {
	return os.WriteFile(file, []byte(source), 0644)
}

func (c *Cli) createClient(config client.Config) client.Client {
	return client.NewClient(config)
}

func createLogger() *zap.Logger {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "M",
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger := zap.Must(cfg.Build())
	return logger
}
