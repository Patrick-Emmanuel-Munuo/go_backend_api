# Step 1: Define project path
$projectPath = "C:\Users\Eng VarTrick\Documents\GitHub\go_backend_api\configurations"

# Ensure the output directory exists
New-Item -ItemType Directory -Path $projectPath -Force

# Step 2: Generate the self-signed certificate for 192.168.137.1
$cert = New-SelfSignedCertificate `
    -Subject 'CN=192.168.137.1' `
    -DnsName '192.168.137.1' `
    -CertStoreLocation 'Cert:\LocalMachine\My' `
    -KeyLength 2048 `
    -NotAfter (Get-Date).AddYears(2) `
    -KeyExportPolicy Exportable

# Step 3: Secure the password for export
$pwd = ConvertTo-SecureString -String 'Variable98@2025' -Force -AsPlainText

# Step 4: Export the certificate as PFX into your project folder
Export-PfxCertificate `
    -Cert "Cert:\LocalMachine\My\$($cert.Thumbprint)" `
    -FilePath "$projectPath\server.pfx" `
    -Password $pwd

# Step 5: Convert PFX to PEM files (cert + key)
openssl pkcs12 -in "$projectPath\server.pfx" -clcerts -nokeys -out "$projectPath\cert.pem" -passin pass:Variable98@2025
openssl pkcs12 -in "$projectPath\server.pfx" -nocerts -nodes -out "$projectPath\key.pem" -passin pass:Variable98@2025
