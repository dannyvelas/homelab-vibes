package wireguard

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/homelab-vibe/iac/config"
)

// Peer represents a WireGuard peer configuration.
type Peer struct {
	Name       string
	PublicKey  string
	PrivateKey string
	PresharedKey string
	AllowedIPs string
	Endpoint   string
}

// GenerateKeys generates a WireGuard key pair and preshared key.
func GenerateKeys() (privateKey, publicKey, presharedKey string, err error) {
	// Generate private key
	privOut, err := exec.Command("wg", "genkey").Output()
	if err != nil {
		return "", "", "", fmt.Errorf("generating private key: %w", err)
	}
	privateKey = strings.TrimSpace(string(privOut))

	// Derive public key
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privateKey)
	pubOut, err := cmd.Output()
	if err != nil {
		return "", "", "", fmt.Errorf("generating public key: %w", err)
	}
	publicKey = strings.TrimSpace(string(pubOut))

	// Generate preshared key
	pskOut, err := exec.Command("wg", "genpsk").Output()
	if err != nil {
		return "", "", "", fmt.Errorf("generating preshared key: %w", err)
	}
	presharedKey = strings.TrimSpace(string(pskOut))

	return privateKey, publicKey, presharedKey, nil
}

// AllocateIP assigns the next available VPN IP from the subnet.
// It reads existing peers from the server config to find the next free IP.
func AllocateIP(cfg *config.Config, serverHost string) (string, error) {
	host := cfg.Hosts[serverHost]

	// Try to read existing peers via SSH
	out, err := exec.Command("ssh", "-o", "ConnectTimeout=5",
		fmt.Sprintf("%s@%s", cfg.Secrets.SSHUser, host.IP),
		"wg show wg0 allowed-ips",
	).Output()

	usedIPs := make(map[string]bool)
	usedIPs["10.0.0.1"] = true // Server IP is always taken

	if err == nil {
		// Parse "peer_pubkey\t10.0.0.X/32" lines
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				ip := strings.TrimSuffix(fields[1], "/32")
				usedIPs[ip] = true
			}
		}
	}

	// Find next available IP (start from .2)
	for i := 2; i < 255; i++ {
		candidate := fmt.Sprintf("10.0.0.%d", i)
		if !usedIPs[candidate] {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("no available IPs in VPN subnet")
}

// GenerateClientConfig creates a WireGuard client configuration file.
func GenerateClientConfig(peer Peer, cfg *config.Config, serverPublicKey string) string {
	var sb strings.Builder

	sb.WriteString("[Interface]\n")
	sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", peer.PrivateKey))
	sb.WriteString(fmt.Sprintf("Address = %s/32\n", peer.AllowedIPs))
	sb.WriteString("DNS = 1.1.1.1, 8.8.8.8\n")
	sb.WriteString("\n")
	sb.WriteString("[Peer]\n")
	sb.WriteString(fmt.Sprintf("PublicKey = %s\n", serverPublicKey))
	sb.WriteString(fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey))
	sb.WriteString(fmt.Sprintf("Endpoint = %s:%d\n", cfg.VPN.Endpoint, cfg.VPN.Port))

	// AllowedIPs: VPN subnet + any LAN routes
	allowedIPs := []string{cfg.VPN.Subnet}
	allowedIPs = append(allowedIPs, cfg.VPN.LANRoutes...)
	sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", strings.Join(allowedIPs, ", ")))

	sb.WriteString("PersistentKeepalive = 25\n")

	return sb.String()
}

// AddPeerToServer adds a peer to the WireGuard server via SSH.
func AddPeerToServer(cfg *config.Config, serverHost string, peer Peer) error {
	host := cfg.Hosts[serverHost]

	// Add peer using wg set command
	cmd := fmt.Sprintf("wg set wg0 peer %s preshared-key <(echo '%s') allowed-ips %s/32",
		peer.PublicKey, peer.PresharedKey, peer.AllowedIPs)

	if err := exec.Command("ssh", "-o", "ConnectTimeout=5",
		fmt.Sprintf("%s@%s", cfg.Secrets.SSHUser, host.IP),
		"bash", "-c", cmd,
	).Run(); err != nil {
		return fmt.Errorf("adding peer to server: %w", err)
	}

	// Save config
	if err := exec.Command("ssh", "-o", "ConnectTimeout=5",
		fmt.Sprintf("%s@%s", cfg.Secrets.SSHUser, host.IP),
		"wg-quick save wg0",
	).Run(); err != nil {
		return fmt.Errorf("saving server config: %w", err)
	}

	return nil
}

// RemovePeer removes a peer from the WireGuard server via SSH.
func RemovePeer(cfg *config.Config, serverHost string, publicKey string) error {
	host := cfg.Hosts[serverHost]

	if err := exec.Command("ssh", "-o", "ConnectTimeout=5",
		fmt.Sprintf("%s@%s", cfg.Secrets.SSHUser, host.IP),
		"wg", "set", "wg0", "peer", publicKey, "remove",
	).Run(); err != nil {
		return fmt.Errorf("removing peer: %w", err)
	}

	// Save config
	if err := exec.Command("ssh", "-o", "ConnectTimeout=5",
		fmt.Sprintf("%s@%s", cfg.Secrets.SSHUser, host.IP),
		"wg-quick", "save", "wg0",
	).Run(); err != nil {
		return fmt.Errorf("saving server config: %w", err)
	}

	return nil
}

// ListPeers returns a list of peers from the WireGuard server.
func ListPeers(cfg *config.Config, serverHost string) ([]string, error) {
	host := cfg.Hosts[serverHost]

	out, err := exec.Command("ssh", "-o", "ConnectTimeout=5",
		fmt.Sprintf("%s@%s", cfg.Secrets.SSHUser, host.IP),
		"wg", "show", "wg0", "peers",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("listing peers: %w", err)
	}

	var peers []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			peers = append(peers, line)
		}
	}

	return peers, nil
}

// SaveClientConfig writes the client config to a file and optionally generates a QR code.
func SaveClientConfig(outputDir, peerName, configContent string) (string, error) {
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return "", fmt.Errorf("creating output directory: %w", err)
	}

	configFile := filepath.Join(outputDir, fmt.Sprintf("%s.conf", peerName))
	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		return "", fmt.Errorf("writing config file: %w", err)
	}

	// Try to generate QR code (qrencode is optional)
	qrFile := filepath.Join(outputDir, fmt.Sprintf("%s.png", peerName))
	qrCmd := exec.Command("qrencode", "-t", "png", "-o", qrFile, configContent)
	if err := qrCmd.Run(); err != nil {
		// QR code generation is optional
		fmt.Println("  (qrencode not installed — skipping QR code generation)")
	} else {
		fmt.Printf("  QR code: %s\n", qrFile)
	}

	return configFile, nil
}

// GetServerPublicKey retrieves the server's public key via SSH.
func GetServerPublicKey(cfg *config.Config, serverHost string) (string, error) {
	host := cfg.Hosts[serverHost]

	out, err := exec.Command("ssh", "-o", "ConnectTimeout=5",
		fmt.Sprintf("%s@%s", cfg.Secrets.SSHUser, host.IP),
		"cat", "/etc/wireguard/keys/server_public.key",
	).Output()
	if err != nil {
		return "", fmt.Errorf("reading server public key: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

// base64Encode is a helper for encoding strings.
func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
