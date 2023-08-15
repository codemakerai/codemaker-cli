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
	"strings"
	"time"
)

const (
	initialRetryDelay  = 1 * time.Second
	maxRetryDelay      = 60 * time.Second
	processTimeout     = 10 * time.Minute
	nonExponentRetries = 8
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
	case "migrate":
		c.parseMigrateArgs()
		break
	case "refactor":
		c.parseRefactorArgs()
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
	case "code":
		generateDocsCmd := flag.NewFlagSet("generateCode", flag.ExitOnError)
		lang := generateDocsCmd.String("language", "", "Programming language: JavaScript, Java, Kotlin")
		replace := generateDocsCmd.Bool("replace", false, "Determines if the existing code is replaced")
		codePath := generateDocsCmd.String("codepath", "", "The codepath to match.")

		err := generateDocsCmd.Parse(os.Args[3:])
		if err != nil {
			c.logger.Errorf("Could not parse args %v", err)
			os.Exit(1)
		}

		if len(generateDocsCmd.Args()) == 0 {
			c.logger.Errorf("Expected file input")
			fmt.Printf("Usage: codemaker generate code <file>\n")
			os.Exit(1)
		}

		config, err := createConfig()
		if err != nil {
			c.logger.Errorf("No valid api key found: %v", err)
			os.Exit(1)
		}

		cl := c.createClient(*config)
		files := generateDocsCmd.Args()[0:]

		if err := c.generateCode(cl, lang, replace, codePath, files); err != nil {
			c.logger.Errorf("Could not generate the code %v", err)
		}
		break
	case "docs":
		generateDocsCmd := flag.NewFlagSet("generateDocs", flag.ExitOnError)
		lang := generateDocsCmd.String("language", "", "Programming language: JavaScript, Java, Kotlin")
		replace := generateDocsCmd.Bool("replace", false, "Determines if the existing documentations are replaced")
		codePath := generateDocsCmd.String("codepath", "", "The codepath to match.")

		err := generateDocsCmd.Parse(os.Args[3:])
		if err != nil {
			c.logger.Errorf("Could not parse args %v", err)
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
		files := generateDocsCmd.Args()[0:]

		if err := c.generateDocumentation(cl, lang, replace, codePath, files); err != nil {
			c.logger.Errorf("Could not generate the documentation %v", err)
		}
		break
	case "unit-tests":
		generateTestsCmd := flag.NewFlagSet("generateUnitTests", flag.ExitOnError)
		lang := generateTestsCmd.String("language", "", "Programming language: JavaScript, Java, Kotlin")
		outputDir := generateTestsCmd.String("output-dir", "", "The output directory")

		err := generateTestsCmd.Parse(os.Args[3:])
		if err != nil {
			c.logger.Error("Could not parse args %v", err)
			os.Exit(1)
		}

		if len(generateTestsCmd.Args()) == 0 {
			c.logger.Errorf("Expected file input")
			fmt.Printf("Usage: codemaker generate tests <file>")
			os.Exit(1)
		}

		config, err := createConfig()
		if err != nil {
			c.logger.Error("No valid api key found %v", err)
			os.Exit(1)
		}

		cl := c.createClient(*config)
		input := generateTestsCmd.Args()[0:]

		if err := c.generateTests(cl, lang, input, outputDir); err != nil {
			c.logger.Errorf("Could not generate the test %v", err)
		}
		break
	default:
		fmt.Printf("Unknown command %s\n", os.Args[2])
		c.printGenerateHelp()
	}
}

func (c *Cli) parseMigrateArgs() {
	if len(os.Args) < 3 {
		fmt.Printf("No command specified")
		c.printMigrateHelp()
	}

	switch os.Args[2] {
	case "syntax":
		migrateSyntaxCmd := flag.NewFlagSet("migrateSyntax", flag.ExitOnError)
		lang := migrateSyntaxCmd.String("language", "", "Programming language: Java")
		langVer := migrateSyntaxCmd.String("language-version", "", "Programming language version")

		err := migrateSyntaxCmd.Parse(os.Args[3:])
		if err != nil {
			c.logger.Errorf("Could not parse args %v", err)
		}

		if len(migrateSyntaxCmd.Args()) == 0 {
			c.logger.Errorf("Expected file input")
			fmt.Printf("Usage: codemaker migrate syntax <file>")
			os.Exit(1)
		}

		config, err := createConfig()
		if err != nil {
			c.logger.Errorf("No valid api key found %v", err)
			os.Exit(1)
		}

		cl := c.createClient(*config)
		input := migrateSyntaxCmd.Args()[0:]

		if err := c.migrateSyntax(cl, lang, langVer, input); err != nil {
			c.logger.Errorf("Could not migrate the syntax %v", err)
		}
		break
	default:
		fmt.Printf("Unknown command %s\n", os.Args[2])
		c.printMigrateHelp()
	}
}

func (c *Cli) parseRefactorArgs() {
	if len(os.Args) < 3 {
		fmt.Printf("No command specified")
		c.printRefactorHelp()
	}

	switch os.Args[2] {
	case "naming":
		refactorNaming := flag.NewFlagSet("refactorNaming", flag.ExitOnError)
		lang := refactorNaming.String("language", "", "Programming language: JavaScript, Java, Kotlin")

		err := refactorNaming.Parse(os.Args[3:])
		if err != nil {
			c.logger.Errorf("Could not parse args %v", err)
		}

		if len(refactorNaming.Args()) == 0 {
			c.logger.Errorf("Expected file input")
			fmt.Printf("Usage: codemaker refactor naming <file>")
			os.Exit(1)
		}

		config, err := createConfig()
		if err != nil {
			c.logger.Errorf("No valid api key found %v", err)
			os.Exit(1)
		}

		cl := c.createClient(*config)
		input := refactorNaming.Args()[0:]

		if err := c.refactorNaming(cl, lang, input); err != nil {
			c.logger.Errorf("Could not rename variables %v", err)
		}
		break
	default:
		fmt.Printf("Unknown command %s\n", os.Args[2])
		c.printRefactorHelp()
	}
}

func (c *Cli) generateCode(cl client.Client, lang *string, replace *bool, codePath *string, files []string) error {
	return c.walkPath(files, func(file string) error {
		if lang == nil || len(*lang) == 0 {
			actLang, err := languageFromExtension(filepath.Ext(file))
			if err != nil {
				return err
			}
			lang = &actLang
		}

		c.logger.Infof("Generating code in file %s", file)
		source, err := c.readFile(file)
		if err != nil {
			return err
		}

		output, err := c.process(cl, client.ModeCode, *lang, *replace, codePath, "", source)
		if err != nil {
			return err
		}

		if err := c.writeFile(file, *output); err != nil {
			return err
		}

		return nil
	})
}

func (c *Cli) generateDocumentation(cl client.Client, lang *string, replace *bool, codePath *string, files []string) error {
	return c.walkPath(files, func(file string) error {
		if lang == nil || len(*lang) == 0 {
			actLang, err := languageFromExtension(filepath.Ext(file))
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

		output, err := c.process(cl, client.ModeDocument, *lang, *replace, codePath, "", source)
		if err != nil {
			return err
		}

		if err := c.writeFile(file, *output); err != nil {
			return err
		}

		return nil
	})
}

func (c *Cli) generateTests(cl client.Client, lang *string, files []string, outputDir *string) error {
	return c.walkPath(files, func(file string) error {
		if lang == nil || len(*lang) == 0 {
			actLang, err := languageFromExtension(filepath.Ext(file))
			if err != nil {
				c.logger.Errorf("skipping unsupported file %s", file)
				return err
			}
			lang = &actLang
		}

		c.logger.Infof("Generating tests for file %s", file)
		source, err := c.readFile(file)
		if err != nil {
			c.logger.Errorf("failed to read file %s %v", file, err)
			return err
		}

		output, err := c.process(cl, client.ModeUnitTest, *lang, false, nil, "", source)
		if err != nil {
			c.logger.Errorf("failed to generate documentation for file %s %v", file, err)
			return err
		}

		suffix, err := testFileSuffix(*lang)
		if err != nil {
			c.logger.Errorf("could not get suffix for file %s %v", file, err)
			return err
		}

		var outputFile string
		if outputDir != nil && len(*outputDir) > 0 {
			err := os.MkdirAll(*outputDir, 0755)
			if err != nil {
				c.logger.Errorf("could not create directory %s", *outputDir)
				return err
			}
			outputFile = filepath.Join(*outputDir, strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))+suffix)
		} else {
			outputFile = strings.TrimSuffix(file, filepath.Ext(file)) + suffix
		}
		if err := c.writeFile(outputFile, *output); err != nil {
			c.logger.Errorf("failed to write file %s %v", file, err)
			return err
		}
		return nil
	})
}

func (c *Cli) migrateSyntax(cl client.Client, lang *string, langVer *string, files []string) error {
	return c.walkPath(files, func(file string) error {
		if lang == nil || len(*lang) == 0 {
			actLang, err := languageFromExtension(filepath.Ext(file))
			if err != nil {
				c.logger.Errorf("skipping unsupported file %s", file)
				return nil
			}
			lang = &actLang
		}

		c.logger.Infof("Migrating syntax in file %s", file)
		source, err := c.readFile(file)
		if err != nil {
			c.logger.Errorf("failed to read file %s %v", file, err)
			return nil
		}

		output, err := c.process(cl, client.ModeMigrateSyntax, *lang, false, nil, *langVer, source)
		if err != nil {
			c.logger.Errorf("failed to migrate syntax in file %s %v", file, err)
			return nil
		}

		if err := c.writeFile(file, *output); err != nil {
			c.logger.Errorf("failed to write file %s %v", file, err)
			return nil
		}
		return nil
	})
}

func (c *Cli) refactorNaming(cl client.Client, lang *string, files []string) error {
	return c.walkPath(files, func(file string) error {
		if lang == nil || len(*lang) == 0 {
			actLang, err := languageFromExtension(filepath.Ext(file))
			if err != nil {
				c.logger.Errorf("skipping unsupported file %s", file)
				return nil
			}
			lang = &actLang
		}

		c.logger.Infof("Renaming local variables in file %s", file)
		source, err := c.readFile(file)
		if err != nil {
			c.logger.Errorf("failed to read file %s %v", file, err)
			return nil
		}

		output, err := c.process(cl, client.ModeRefactorNaming, *lang, false, nil, "", source)
		if err != nil {
			c.logger.Errorf("failed to rename variables in file %s %v", file, err)
			return nil
		}

		if err := c.writeFile(file, *output); err != nil {
			c.logger.Errorf("failed to write file %s %v", file, err)
			return nil
		}
		return nil
	})
}

func (c *Cli) process(cl client.Client, mode string, lang string, replace bool, codePath *string, langVer string, source string) (*string, error) {
	modify := client.ModifyNone
	if replace {
		modify = client.ModifyReplace
	}

	if codePath == nil || len(*codePath) == 0 {
		codePath = nil
	}

	process, err := cl.CreateProcess(&client.CreateProcessRequest{
		Process: client.Process{
			Mode:     mode,
			Language: lang,
			Input: client.Input{
				Source: source,
			},
			Options: &client.Options{
				LanguageVersion: &langVer,
				Modify:          &modify,
				CodePath:        codePath,
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
		} else if c.isFailed(status) {
			return nil, fmt.Errorf("the task processing has failed")
		}

		select {
		case <-timeout:
			return nil, fmt.Errorf("the task processing had timed out")
		default:
			c.backoff(retry)
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

func (c *Cli) configure() error {
	c.logger.Infof("Configure CLI")

	var apiKey string

	fmt.Print("Enter API Key: ")
	_, err := fmt.Scanln(&apiKey)
	if err != nil {
		c.logger.Errorf("Failed to read the stdin %v", err)
		return err
	}

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

func (c *Cli) isCompleted(status *client.GetProcessStatusResponse) bool {
	return status.Status == client.StatusCompleted
}

func (c *Cli) isFailed(status *client.GetProcessStatusResponse) bool {
	return status.Status == client.StatusFailed ||
		status.Status == client.StatusTimedOut
}

func (c *Cli) backoff(retry int) {
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

func (c *Cli) printHelp() {
	fmt.Printf("Usage: codemaker <command>\n")
	fmt.Printf("\n")
	fmt.Printf("Commands:\n")
	fmt.Printf(" * generate\n")
	fmt.Printf(" * migrate\n")
	fmt.Printf(" * refactor\n")
	fmt.Printf(" * configure\n")
	fmt.Printf(" * version\n")
	os.Exit(1)
}

func (c *Cli) printGenerateHelp() {
	fmt.Printf("Usage: codemaker generate <command>\n")
	fmt.Printf("\n")
	fmt.Printf("Commands:\n")
	fmt.Printf(" * code\n")
	fmt.Printf(" * docs\n")
	fmt.Printf(" * unit-tests\n")
	os.Exit(1)
}

func (c *Cli) printMigrateHelp() {
	fmt.Printf("Usage: codemaker migrate <command>\n")
	fmt.Printf("\n")
	fmt.Printf("Commands:\n")
	fmt.Printf(" * syntax\n")
	os.Exit(1)
}

func (c *Cli) printRefactorHelp() {
	fmt.Printf("Usage: codemaker refactor <command>\n")
	fmt.Printf("\n")
	fmt.Printf("Commands:\n")
	fmt.Printf(" * naming\n")
	os.Exit(1)
}

func (c *Cli) walkPath(files []string, visitor func(file string) error) error {
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
