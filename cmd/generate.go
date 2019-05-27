package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/zededa/adam/pkg/x509"
)

var (
	cn string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate certs for the Adam server and clients",
	Long:  `Generate the necessary certs for the Adam server and clients`,
}

var generateServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Generate server certs",
	Long:  `Generate the necessary server certs`,
	Run: func(cmd *cobra.Command, args []string) {
		err := x509.Generate("", hosts, certPath, keyPath, force)
		if err != nil {
			log.Fatalf("error generating key/cert: %v", err)
		}
	},
}

var generateOnboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Generate onboarding certs",
	Long:  `Generate an onboarding cert. The cert will be saved in the provided path, named by the CN, e.g. onboard/company-a.pem and onboard/company-a-key.pem.`,
	Run: func(cmd *cobra.Command, args []string) {
		if onboardingDatabasePath == "" {
			log.Fatalf("onboarding path must be set")
		}
		fi, err := os.Stat(onboardingDatabasePath)
		if err != nil {
			log.Fatalf("onboarding database path %s does not exist", onboardingDatabasePath)
		}
		if !fi.IsDir() {
			log.Fatalf("onboarding database path %s is not a directory", onboardingDatabasePath)
		}
		re := regexp.MustCompile(`[^a-zA-Z0-9\\.\\-]`)
		cnSquashed := re.ReplaceAllString(cn, "_")
		certPath := path.Join(onboardingDatabasePath, fmt.Sprintf("%s.pem", cnSquashed))
		keyPath := path.Join(onboardingDatabasePath, fmt.Sprintf("%s-key.pem", cnSquashed))
		err = x509.Generate(cn, "", certPath, keyPath, force)
		if err != nil {
			log.Fatalf("error generating key/cert: %v", err)
		}
	},
}

var generateDeviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Generate individual device certs",
	Long:  `Generate a device cert. The cert will be saved in the provided path, named by the CN, e.g. onboard/device-1234.pem and onboard/device-1234-key.pem.`,
	Run: func(cmd *cobra.Command, args []string) {
		if deviceDatabasePath == "" {
			log.Fatalf("device path must be set")
		}
		fi, err := os.Stat(deviceDatabasePath)
		if err != nil {
			log.Fatalf("device database path %s does not exist", deviceDatabasePath)
		}
		if !fi.IsDir() {
			log.Fatalf("device database path %s is not a directory", deviceDatabasePath)
		}
		re := regexp.MustCompile(`[^a-zA-Z0-9\\.\\-]`)
		cnSquashed := re.ReplaceAllString(cn, "_")
		certPath := path.Join(deviceDatabasePath, fmt.Sprintf("%s.pem", cnSquashed))
		keyPath := path.Join(deviceDatabasePath, fmt.Sprintf("%s-key.pem", cnSquashed))
		err = x509.Generate(cn, "", certPath, keyPath, force)
		if err != nil {
			log.Fatalf("error generating key/cert: %v", err)
		}
	},
}

func generateInit() {
	// generate server
	generateCmd.AddCommand(generateServerCmd)
	generateServerCmd.Flags().StringVar(&certPath, "certfile", defaultCertPath, "path to server certificate")
	generateServerCmd.Flags().StringVar(&keyPath, "keyfile", defaultKeyPath, "path to server key")
	generateServerCmd.Flags().StringVar(&hosts, "hosts", "", "hostnames and/or IPs to use in the certificate, separated by ',', output to the certfile and keyfile; will not replace if they exist")
	generateServerCmd.MarkFlagRequired("hosts")
	generateServerCmd.Flags().BoolVar(&force, "force", false, "replace existing files")

	// generate onboarding certs
	generateCmd.AddCommand(generateOnboardCmd)
	generateOnboardCmd.Flags().StringVar(&onboardingDatabasePath, "onboard-db", "", "path to directory where we will store the generated onboarding certificates")
	generateOnboardCmd.MarkFlagRequired("onboard-db")
	generateOnboardCmd.Flags().StringVar(&cn, "cn", "", "CN to use in the certificate; will not replace if one with the same CN exists")
	generateOnboardCmd.MarkFlagRequired("cn")
	generateOnboardCmd.Flags().BoolVar(&force, "force", false, "replace existing files")

	// generate device certs
	generateCmd.AddCommand(generateDeviceCmd)
	generateDeviceCmd.Flags().StringVar(&deviceDatabasePath, "device-db", "", "path to directory where we will store the generated device certificates")
	generateDeviceCmd.MarkFlagRequired("device-db")
	generateDeviceCmd.Flags().StringVar(&cn, "cn", "", "CN to use in the certificate; will not replace if one with the same CN exists")
	generateDeviceCmd.MarkFlagRequired("cn")
	generateDeviceCmd.Flags().BoolVar(&force, "force", false, "replace existing files")
}
