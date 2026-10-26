package main

const (
	toolName         = "openssl"
	binName          = "openssl"
	envVar           = "OPENSSL"
	listNoun         = "tasks"
	argsDesc         = "extra arguments (positional values such as files to hash)"
	confirmMsg       = "This may connect to a host or write files. Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "version -a   (tab completes subcommands and flags, enter runs)"
	shellTry         = "version   x509 -in cert.pem -noout -dates   rand -hex 16   dgst -sha256 file   ciphers -v"
	shellNote        = "Reading, hashing and printing run straight away. Anything that connects (s_client,\ns_server, s_time...) or writes a file (-out, -keyout) asks first."
	shellUsage       = "usage: openssl subcommand [options] [args]   tab completes subcommands and flags"
	generalHelp      = `openssl subcommand [options] [arguments]

openssl is many tools in one: the first word is the subcommand (x509, req,
s_client, dgst, enc, genpkey...) and every task here starts with one. Options
come after it; each option that takes a value gets its own row, and [arguments]
holds any plain trailing words.

Most subcommands read their input from a file (-in) or from stdin. This TUI
has no stdin to give, so name files with -in. Output goes to stdout unless
you pass -out, which writes (and overwrites) a file and asks first.

Subcommands that connect to a host (s_client, s_time, ocsp) or listen
(s_server) always ask first. s_server waits for connections, so it runs in the
terminal (t). Use ctrl+c to stop it.

Values for -subj look like "/CN=example.com/O=Example"; quote them in the
shell screen or arguments box: -subj '/CN=example.com'.`
)

var versionArgs = []string{"version"}
var attachedFlags = map[string]bool{}

const (
	gCert   = "Certificates (x509)"
	gRemote = "Remote servers (s_client)"
	gKeys   = "Keys and requests"
	gHash   = "Hashes, encoding and random"
	gInfo   = "What this openssl supports"
)

var (
	fIn     = Flag{"-in", "file", "Input file"}
	fOut    = Flag{"-out", "file", "Write the output to this file (overwrites it)"}
	fNoout  = Flag{"-noout", "", "Do not print the encoded certificate or key itself"}
	fText   = Flag{"-text", "", "Print in human-readable text form"}
	fConn   = Flag{"-connect", "host:port", "Host and port to connect to, e.g. example.com:443"}
	fSNI    = Flag{"-servername", "name", "Server name to send (SNI); needed for hosts sharing an address"}
	fSubj   = Flag{"-subj", "/CN=name/O=org", "Subject for the request or certificate (avoids the questions)"}
	fKey    = Flag{"-key", "file", "Private key file"}
	fKeyout = Flag{"-keyout", "file", "File to write the new private key to"}
	fDays   = Flag{"-days", "n", "Days the certificate is valid"}
	fCA     = Flag{"-CAfile", "file", "File of trusted CA certificates to check against"}
)

var Cmds = []Cmd{
	{Name: "cert-info", Group: gCert, Prefix: []string{"x509"},
		Usage:   "openssl x509 -in cert.pem -noout -text",
		Desc:    "Decode a PEM certificate and print everything in it: subject, issuer, validity dates, key, extensions (SANs, key usage), signature.",
		Example: "openssl x509 -in cert.pem -noout -text",
		Flags:   []Flag{fIn, fNoout, fText, {"-subject", "", "Print the subject name"}, {"-issuer", "", "Print the issuer name"}, {"-dates", "", "Print notBefore and notAfter"}, {"-fingerprint", "", "Print the certificate fingerprint"}, {"-serial", "", "Print the serial number"}}},
	{Name: "cert-dates", Group: gCert, Prefix: []string{"x509"},
		Usage:   "openssl x509 -in cert.pem -noout -dates",
		Desc:    "When does a certificate start and stop being valid? With -checkend N the exit status says whether it expires within N seconds (86400 = a day): 0 is fine, 1 is expiring.",
		Example: "openssl x509 -in cert.pem -noout -dates",
		Flags:   []Flag{fIn, fNoout, {"-dates", "", "Print notBefore and notAfter"}, {"-enddate", "", "Print only notAfter"}, {"-checkend", "seconds", "Exit 1 if the certificate expires within this many seconds"}}},
	{Name: "cert-names", Group: gCert, Prefix: []string{"x509"},
		Usage:   "openssl x509 -in cert.pem -noout -subject -issuer -ext subjectAltName",
		Desc:    "Who is a certificate for and who signed it? -ext subjectAltName prints the hostnames it covers, which is what browsers check.",
		Example: "openssl x509 -in cert.pem -noout -subject -issuer -ext subjectAltName",
		Flags:   []Flag{fIn, fNoout, {"-subject", "", "Print the subject name"}, {"-issuer", "", "Print the issuer name"}, {"-ext", "extension", "Print only these extensions, e.g. subjectAltName"}, {"-nameopt", "option", "Name format, e.g. RFC2253 or multiline"}}},
	{Name: "verify-chain", Group: gCert, Prefix: []string{"verify"},
		Usage:   "openssl verify -CAfile ca.pem [-untrusted chain.pem] cert.pem",
		Desc:    "Check that a certificate chains to a trusted CA. Put the certificate file in [arguments]; give intermediates with -untrusted.",
		Example: "openssl verify -CAfile ca.pem cert.pem",
		Flags:   []Flag{fCA, {"-untrusted", "file", "File of intermediate certificates (not trusted themselves)"}, {"-purpose", "purpose", "Required purpose, e.g. sslserver"}, {"-show_chain", "", "Print the chain that was built"}}},

	{Name: "remote-cert", Group: gRemote, Prefix: []string{"s_client"},
		Usage:   "openssl s_client -connect host:443 -servername host",
		Desc:    "Do a TLS handshake with a server and print what it presented: the certificate chain, negotiated protocol and cipher, and the verification result (`Verify return code: 0 (ok)`). It reads no input here, so it disconnects right away. It connects to a remote host, so it asks first.",
		Example: "openssl s_client -connect example.com:443 -servername example.com",
		Flags:   []Flag{fConn, fSNI, {"-brief", "", "Just a short summary of the connection"}, {"-showcerts", "", "Print every certificate in the chain, in PEM"}, fCA, {"-verify_hostname", "name", "Check the certificate matches this hostname"}, {"-status", "", "Ask for an OCSP-stapled response"}}},
	{Name: "tls-version", Group: gRemote, Prefix: []string{"s_client"},
		Usage:   "openssl s_client -connect host:443 -tls1_2 | -tls1_3 | -no_tls1_3",
		Desc:    "Test which TLS version a server will speak by forcing one: a failed handshake means it is not accepted. -brief keeps the output short.",
		Example: "openssl s_client -connect example.com:443 -tls1_2 -brief",
		Flags:   []Flag{fConn, fSNI, {"-tls1_2", "", "Only TLS 1.2"}, {"-tls1_3", "", "Only TLS 1.3"}, {"-no_tls1_3", "", "Do not offer TLS 1.3"}, {"-brief", "", "Just a short summary of the connection"}, {"-cipher", "list", "Cipher list for TLS 1.2 and below"}, {"-alpn", "protocols", "Offer ALPN protocols, e.g. h2,http/1.1"}}},
	{Name: "starttls", Group: gRemote, Prefix: []string{"s_client"},
		Usage: "openssl s_client -connect host:25 -starttls smtp",
		Desc:  "Check TLS on a service that begins in plain text and upgrades: smtp, imap, pop3, ftp, xmpp, postgres and others.",
		Flags: []Flag{fConn, {"-starttls", "protocol", "smtp, pop3, imap, ftp, xmpp, postgres, mysql, ..."}, fSNI, {"-brief", "", "Just a short summary of the connection"}, {"-showcerts", "", "Print every certificate in the chain, in PEM"}}},
	{Name: "serve", Group: gRemote, Prefix: []string{"s_server"},
		Usage: "openssl s_server -accept port -cert cert.pem -key key.pem",
		Desc:  "Run a test TLS server that answers connections on a port, to try a client against. It waits for connections, so it runs in the terminal (t) and stops with ctrl+c. Asks first.",
		Flags: []Flag{{"-accept", "port", "Port to listen on"}, {"-cert", "file", "Certificate file"}, fKey, {"-www", "", "Send a status page to clients"}, {"-tls1_3", "", "Only TLS 1.3"}}},

	{Name: "gen-key", Group: gKeys, Prefix: []string{"genpkey"},
		Usage:   "openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out key.pem",
		Desc:    "Generate a private key: RSA (with the bit size), EC (with ec_paramgen_curve:P-256) or ED25519. Without -out the key is printed here. It writes a file with -out, so that asks first.",
		Example: "openssl genpkey -algorithm ED25519",
		Flags:   []Flag{{"-algorithm", "alg", "RSA, EC, ED25519, X25519 ..."}, {"-pkeyopt", "opt:value", "Key option, e.g. rsa_keygen_bits:4096 or ec_paramgen_curve:P-256"}, fOut, {"-aes-256-cbc", "", "Encrypt the key with a passphrase"}, fText}},
	{Name: "self-signed", Group: gKeys, Prefix: []string{"req"},
		Usage:   "openssl req -x509 -newkey rsa:2048 -nodes -keyout key.pem -out cert.pem -days 365 -subj /CN=localhost",
		Desc:    "Make a new key and a self-signed certificate in one step, for testing. -noenc leaves the key unencrypted. Writes two files, so it asks first. Add -addext subjectAltName=DNS:localhost for browsers.",
		Example: "openssl req -x509 -newkey rsa:2048 -noenc -keyout key.pem -out cert.pem -days 365 -subj /CN=localhost",
		Flags:   []Flag{{"-x509", "", "Output a certificate rather than a request"}, {"-newkey", "alg:bits", "Generate a new key, e.g. rsa:2048 or ec:param.pem"}, {"-noenc", "", "Do not encrypt the private key"}, fKeyout, fOut, fDays, fSubj, {"-addext", "key=value", "Add an extension, e.g. subjectAltName=DNS:localhost"}}},
	{Name: "csr", Group: gKeys, Prefix: []string{"req"},
		Usage:   "openssl req -new -key key.pem -out req.csr -subj /CN=example.com",
		Desc:    "Create a certificate signing request from an existing key, to send to a CA. Add -addext subjectAltName=... for the hostnames.",
		Example: "openssl req -new -key key.pem -out req.csr -subj /CN=example.com",
		Flags:   []Flag{{"-new", "", "Make a new request"}, fKey, fOut, fSubj, {"-addext", "key=value", "Add an extension, e.g. subjectAltName=DNS:example.com"}}},
	{Name: "csr-info", Group: gKeys, Prefix: []string{"req"},
		Usage: "openssl req -in req.csr -noout -text -verify",
		Desc:  "Read a certificate signing request: its subject, key and requested extensions; -verify checks its signature.",
		Flags: []Flag{fIn, fNoout, fText, {"-verify", "", "Verify the request's signature"}, {"-subject", "", "Print the subject"}}},
	{Name: "key-info", Group: gKeys, Prefix: []string{"pkey"},
		Usage:   "openssl pkey -in key.pem -noout -text",
		Desc:    "Show a private or public key's parameters. -pubout prints the matching public key; -pubin reads a public key.",
		Example: "openssl pkey -in key.pem -noout -text",
		Flags:   []Flag{fIn, fNoout, fText, {"-pubout", "", "Print only the public key"}, {"-pubin", "", "The input is a public key"}}},

	{Name: "digest", Group: gHash, Prefix: []string{"dgst"},
		Usage:   "openssl dgst -sha256 file",
		Desc:    "Hash a file: SHA-256, SHA-512, SHA3-256, MD5 and others. Put the file in [arguments]. Check a download against its published checksum.",
		Example: "openssl dgst -sha256 file.bin",
		Flags:   []Flag{{"-sha256", "", "SHA-256"}, {"-sha512", "", "SHA-512"}, {"-sha1", "", "SHA-1 (weak)"}, {"-md5", "", "MD5 (broken, for old checksums only)"}, {"-sha3-256", "", "SHA3-256"}, {"-hex", "", "Print as hex"}, {"-r", "", "Print in coreutils sha256sum format"}}},
	{Name: "base64", Group: gHash, Prefix: []string{"enc"},
		Usage:   "openssl enc -base64 [-d] -in file",
		Desc:    "Encode a file to Base64 (add -A to keep it on one line), or decode with -d.",
		Example: "openssl enc -base64 -A -in file.bin",
		Flags:   []Flag{{"-base64", "", "Base64 encode (or decode with -d)"}, {"-d", "", "Decrypt or decode"}, {"-A", "", "Single-line Base64"}, fIn, fOut}},
	{Name: "encrypt", Group: gHash, Prefix: []string{"enc"},
		Usage: "openssl enc -aes-256-cbc -pbkdf2 -salt -in plain -out secret.enc",
		Desc:  "Encrypt or decrypt a file with a passphrase. Always give -pbkdf2 (a strong key derivation); decrypt with -d and the same options. It prompts for a passphrase, so run it in the terminal (t), where it writes the output file.",
		Flags: []Flag{{"-aes-256-cbc", "", "AES-256 in CBC mode"}, {"-pbkdf2", "", "Derive the key with PBKDF2"}, {"-iter", "n", "Iteration count (forces PBKDF2)"}, {"-salt", "", "Use a random salt (default)"}, {"-d", "", "Decrypt"}, fIn, fOut, {"-pass", "source", "Passphrase source, e.g. pass:secret or file:path"}}},
	{Name: "random", Group: gHash, Prefix: []string{"rand"},
		Usage:   "openssl rand -hex 16",
		Desc:    "Cryptographically random bytes for a key, token or password: -hex or -base64 for text, or give the byte count in [arguments] (e.g. 32).",
		Example: "openssl rand -hex 16",
		Flags:   []Flag{{"-hex", "", "Print as hex"}, {"-base64", "", "Print as Base64"}}},

	{Name: "ciphers", Group: gInfo, Prefix: []string{"ciphers"},
		Usage:   "openssl ciphers -v ['cipher list']",
		Desc:    "List the cipher suites this openssl supports, with protocol, key exchange and encryption. Put a cipher string such as 'HIGH' or 'ECDHE+AESGCM' in [arguments] to filter.",
		Example: "openssl ciphers -v",
		Flags:   []Flag{{"-v", "", "Verbose listing"}, {"-V", "", "Even more verbose (with IDs)"}, {"-s", "", "Only the supported ciphers"}, {"-tls1_2", "", "Only those usable in TLS 1.2"}, {"-tls1_3", "", "Only those usable in TLS 1.3"}}},
	{Name: "list", Group: gInfo, Prefix: []string{"list"},
		Usage:   "openssl list -digest-algorithms | -cipher-algorithms | -public-key-algorithms",
		Desc:    "List the digests, ciphers, key types, KDFs or MACs that this build supports.",
		Example: "openssl list -digest-algorithms",
		Flags:   []Flag{{"-digest-algorithms", "", "List the hash algorithms"}, {"-cipher-algorithms", "", "List the symmetric ciphers"}, {"-public-key-algorithms", "", "List the public key algorithms"}, {"-kdf-algorithms", "", "List the key derivation functions"}, {"-mac-algorithms", "", "List the MACs"}, {"-providers", "", "List the loaded providers"}}},
	{Name: "version", Group: gInfo, Prefix: []string{"version"},
		Usage:   "openssl version [-a]",
		Desc:    "The version; -a adds build date, platform, compiler flags and directories (where openssl.cnf and the CA store are).",
		Example: "openssl version -a",
		Flags:   []Flag{{"-a", "", "Everything"}, {"-d", "", "OPENSSLDIR: where the config and certs live"}, {"-v", "", "Version only"}, {"-b", "", "Build date"}}},
}

var Popular = map[string]bool{
	"cert-info": true, "cert-dates": true, "remote-cert": true, "self-signed": true,
	"csr": true, "digest": true, "random": true, "tls-version": true, "gen-key": true,
}

var PopularFlags = map[string]bool{"-in": true, "-out": true, "-noout": true, "-text": true, "-connect": true, "-servername": true, "-subj": true, "-days": true, "-key": true}

// Extras are the subcommands, for tab completion.
func Extras() []string {
	return []string{"asn1parse", "ca", "ciphers", "crl", "dgst", "dhparam", "ec", "ecparam", "enc", "genpkey", "genrsa",
		"help", "list", "passwd", "pkcs12", "pkey", "pkeyutl", "rand", "req", "rsa", "s_client", "s_server", "s_time",
		"verify", "version", "x509"}
}

// riskySub are subcommands that contact a host, listen, or modify files.
var riskySub = map[string]bool{"s_client": true, "s_server": true, "s_time": true, "ocsp": true, "ca": true,
	"rehash": true, "c_rehash": true, "cmp": true, "speed": true, "ts": true, "fipsinstall": true, "srp": true}

// writeFlags make openssl write a file.
var writeFlags = []string{"-out", "-keyout", "-writerand", "-outdir", "-CAcreateserial", "-in-place", "-genkey", "-out_dir"}

// IsSafe reports whether argv runs without confirmation: it only reads and
// prints, with no connection and no file written.
func IsSafe(argv []string) bool {
	if len(argv) == 0 {
		return true
	}
	return !riskySub[argv[0]] && !has(argv, writeFlags...)
}

// NeedsTerminal: s_server waits for connections, enc prompts for a passphrase
// unless one is given.
func NeedsTerminal(argv []string) bool {
	if len(argv) == 0 {
		return false
	}
	switch argv[0] {
	case "s_server":
		return true
	case "enc":
		return !has(argv, "-pass", "-k", "-K", "-base64", "-a") && !has(argv, "-md5")
	}
	return false
}
