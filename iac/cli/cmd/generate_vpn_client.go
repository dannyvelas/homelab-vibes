package cmd

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/homelab-vibe/iac/config"
	"github.com/homelab-vibe/iac/wireguard"
)

func generateVPNClientImpl(args []string) error {
	fs := flag.NewFlagSet("generate vpn-client", flag.ExitOnError)
	peerName := fs.String("name", "", "Name for the VPN client/peer")
	configPath := fs.String("config", "homelab.yml", "Path to homelab.yml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *peerName == "" {
		return fmt.Errorf("--name is required. Usage: iac generate vpn-client --name <peer_name>")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Find WireGuard host
	wgHost := cfg.WireGuardHost()
	if wgHost == "" {
		return fmt.Errorf("no host has wireguard_endpoint: true in homelab.yml")
	}

	fmt.Printf("Generating VPN client config for peer: %s\n", *peerName)

	// Generate key pair
	privateKey, publicKey, presharedKey, err := wireguard.GenerateKeys()
	if err != nil {
		return fmt.Errorf("generating keys: %w", err)
	}

	// Allocate VPN IP
	vpnIP, err := wireguard.AllocateIP(cfg, wgHost)
	if err != nil {
		return fmt.Errorf("allocating VPN IP: %w", err)
	}

	peer := wireguard.Peer{
		Name:         *peerName,
		PublicKey:    publicKey,
		PrivateKey:   privateKey,
		PresharedKey: presharedKey,
		AllowedIPs:   vpnIP,
	}

	// Get server public key
	serverPubKey, err := wireguard.GetServerPublicKey(cfg, wgHost)
	if err != nil {
		return fmt.Errorf("getting server public key: %w", err)
	}

	// Generate client config
	clientConfig := wireguard.GenerateClientConfig(peer, cfg, serverPubKey)

	// Add peer to server
	if err := wireguard.AddPeerToServer(cfg, wgHost, peer); err != nil {
		return fmt.Errorf("adding peer to server: %w", err)
	}

	// Save client config
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	outputDir := filepath.Join(repoRoot, ".generated", "vpn-clients")
	configFile, err := wireguard.SaveClientConfig(outputDir, *peerName, clientConfig)
	if err != nil {
		return fmt.Errorf("saving client config: %w", err)
	}

	fmt.Printf("\nVPN client config generated!\n")
	fmt.Printf("  Peer name: %s\n", *peerName)
	fmt.Printf("  VPN IP: %s\n", vpnIP)
	fmt.Printf("  Config file: %s\n", configFile)
	fmt.Printf("  Import into WireGuard client to connect.\n")

	return nil
}
