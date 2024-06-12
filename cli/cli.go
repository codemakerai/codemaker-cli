// Copyright 2023 CodeMaker AI Inc. All rights reserved.

package cli

import (
	"context"
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
	case "fix":
		c.parseFixArgs()
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
		generateCodeCmd := flag.NewFlagSet("generateCode", flag.ExitOnError)
		lang := generateCodeCmd.String("language", "", "Programming language: JavaScript, Java, Kotlin")
		replace := generateCodeCmd.Bool("replace", false, "Determines if the existing code is replaced")
		codePath := generateCodeCmd.String("codepath", "", "The codepath to match.")
		model := generateCodeCmd.String("model", "", "The fine-tuned model name.")
		endpoint := generateCodeCmd.String("endpoint", "", "The endpoint name.")

		err := generateCodeCmd.Parse(os.Args[3:])
		if err != nil {
			c.logger.Errorf("Could not parse args %v", err)
			os.Exit(1)
		}

		if len(generateCodeCmd.Args()) == 0 {
			c.logger.Errorf("Expected file input")
			fmt.Printf("Usage: codemaker generate code <file>\n")
			os.Exit(1)
		}

		config, err := createConfig(endpoint)
		if err != nil {
			c.logger.Errorf("No valid api key found: %v", err)
			os.Exit(1)
		}

		cl, err := c.createClient(*config)
		if err != nil {
			c.logger.Errorf("Can not create client: %v", err)
			os.Exit(1)
		}

		files := generateCodeCmd.Args()[0:]

		if err := c.generateCode(cl, lang, replace, codePath, model, files); err != nil {
			c.logger.Errorf("Could not generate the code %v", err)
		}
		break
	case "docs":
		generateDocsCmd := flag.NewFlagSet("generateDocs", flag.ExitOnError)
		lang := generateDocsCmd.String("language", "", "Programming language: JavaScript, Java, Kotlin")
		replace := generateDocsCmd.Bool("replace", false, "Determines if the existing documentations are replaced")
		codePath := generateDocsCmd.String("codepath", "", "The codepath to match.")
		endpoint := generateDocsCmd.String("endpoint", "", "The endpoint name.")

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

		config, err := createConfig(endpoint)
		if err != nil {
			c.logger.Errorf("No valid api key found: %v", err)
			os.Exit(1)
		}

		cl, err := c.createClient(*config)
		if err != nil {
			c.logger.Errorf("Can not create client: %v", err)
			os.Exit(1)
		}

		files := generateDocsCmd.Args()[0:]

		if err := c.generateDocumentation(cl, lang, replace, codePath, files); err != nil {
			c.logger.Errorf("Could not generate the documentation %v", err)
		}
		break
	default:
		fmt.Printf("Unknown command %s\n", os.Args[2])
		c.printGenerateHelp()
	}
}

func (c *Cli) parseFixArgs() {
	if len(os.Args) < 3 {
		fmt.Printf("No command specified")
		c.printRefactorHelp()
	}

	switch os.Args[2] {
	case "syntax":
		correctSyntaxCmd := flag.NewFlagSet("fixSyntax", flag.ExitOnError)
		lang := correctSyntaxCmd.String("language", "", "Programming language: JavaScript, Java, Kotlin")
		endpoint := correctSyntaxCmd.String("endpoint", "", "The endpoint name.")

		err := correctSyntaxCmd.Parse(os.Args[3:])
		if err != nil {
			c.logger.Errorf("Could not parse args %v", err)
		}

		if len(correctSyntaxCmd.Args()) == 0 {
			c.logger.Errorf("Expected file input")
			fmt.Printf("Usage: codemaker fix syntax <file>")
			os.Exit(1)
		}

		config, err := createConfig(endpoint)
		if err != nil {
			c.logger.Errorf("No valid api key found %v", err)
			os.Exit(1)
		}

		cl, err := c.createClient(*config)
		if err != nil {
			c.logger.Errorf("Can not create client: %v", err)
			os.Exit(1)
		}

		input := correctSyntaxCmd.Args()[0:]

		if err := c.fixSyntax(cl, lang, input); err != nil {
			c.logger.Errorf("Could not fix syntax %v", err)
		}
		break
	default:
		fmt.Printf("Unknown command %s\n", os.Args[2])
		c.printFixHelp()
	}
}

func (c *Cli) generateCode(cl client.Client, lang *string, replace *bool, codePath *string, model *string, files []string) error {
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

		output, err := c.process(cl, client.ModeCode, *lang, *replace, codePath, model, source)
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

		output, err := c.process(cl, client.ModeDocument, *lang, *replace, codePath, nil, source)
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

		output, err := c.process(cl, client.ModeUnitTest, *lang, false, nil, nil, source)
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

func (c *Cli) migrateSyntax(cl client.Client, lang *string, files []string) error {
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

		output, err := c.process(cl, client.ModeMigrateSyntax, *lang, false, nil, nil, source)
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

		output, err := c.process(cl, client.ModeRefactorNaming, *lang, false, nil, nil, source)
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

func (c *Cli) fixSyntax(cl client.Client, lang *string, files []string) error {
	return c.walkPath(files, func(file string) error {
		if lang == nil || len(*lang) == 0 {
			actLang, err := languageFromExtension(filepath.Ext(file))
			if err != nil {
				c.logger.Errorf("skipping unsupported file %s", file)
				return nil
			}
			lang = &actLang
		}

		c.logger.Infof("Fixing syntax in file %s", file)
		source, err := c.readFile(file)
		if err != nil {
			c.logger.Errorf("failed to read file %s %v", file, err)
			return nil
		}

		output, err := c.process(cl, client.ModeFixSyntax, *lang, false, nil, nil, source)
		if err != nil {
			c.logger.Errorf("failed to fix syntax in file %s %v", file, err)
			return nil
		}

		if err := c.writeFile(file, *output); err != nil {
			c.logger.Errorf("failed to write file %s %v", file, err)
			return nil
		}
		return nil
	})
}

func (c *Cli) process(cl client.Client, mode string, lang string, replace bool, codePath *string, model *string, source string) (*string, error) {
	ctx := context.Background()

	modify := client.ModifyNone
	if replace {
		modify = client.ModifyReplace
	}

	if codePath == nil || len(*codePath) == 0 {
		codePath = nil
	}

	process, err := cl.Process(ctx, &client.ProcessRequest{
		Mode:     mode,
		Language: lang,
		Input: client.Input{
			Source: source,
		},
		Options: &client.Options{
			Modify:   &modify,
			CodePath: codePath,
			Model:    model,
		},
	})
	if err != nil {
		return nil, err
	}

	return &process.Source, nil
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
	fmt.Printf(" * refactor\n")
	fmt.Printf(" * fix\n")
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

func (c *Cli) printRefactorHelp() {
	fmt.Printf("Usage: codemaker refactor <command>\n")
	fmt.Printf("\n")
	fmt.Printf("Commands:\n")
	fmt.Printf(" * naming\n")
	os.Exit(1)
}

func (c *Cli) printFixHelp() {
	fmt.Printf("Usage: codemaker fix <command>\n")
	fmt.Printf("\n")
	fmt.Printf("Commands:\n")
	fmt.Printf(" * syntax\n")
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

func (c *Cli) createClient(config client.Config) (client.Client, error) {
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
