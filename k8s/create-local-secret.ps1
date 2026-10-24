[CmdletBinding()]
param(
    [switch] $EnableTicketmaster
)

$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $PSScriptRoot
$envPath = Join-Path $repoRoot 'backend\.env'
$namespace = 'livepulse'
$secretName = 'livepulse-local-secrets'
$expectedContext = 'k3d-livepulse-lab'

if (-not (Test-Path -LiteralPath $envPath)) {
    throw "Missing backend/.env. The local deployment reads the optional Clerk test key from that ignored file."
}

$currentContext = (& kubectl config current-context 2>$null | Out-String).Trim()
if ($LASTEXITCODE -ne 0 -or $currentContext -ne $expectedContext) {
    throw "Refusing to create the Secret. Switch kubectl to '$expectedContext' first; current context is '$currentContext'."
}

$envValues = @{}
foreach ($line in [System.IO.File]::ReadAllLines($envPath)) {
    $trimmed = $line.Trim()
    if (-not $trimmed -or $trimmed.StartsWith('#')) {
        continue
    }

    if ($trimmed -match '^([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$') {
        $key = $Matches[1]
        $value = $Matches[2].Trim()
        if ($value.Length -ge 2) {
            $first = $value[0]
            $last = $value[$value.Length - 1]
            if (($first -eq '"' -and $last -eq '"') -or ($first -eq "'" -and $last -eq "'")) {
                $value = $value.Substring(1, $value.Length - 2)
            }
        }
        $envValues[$key] = $value
    }
}

if ($EnableTicketmaster -and ([string]::IsNullOrWhiteSpace($envValues['EXTERNAL_API_KEY']) -or $envValues['EXTERNAL_API_KEY'] -eq 'your_ticketmaster_api_key')) {
    throw "EXTERNAL_API_KEY is missing or still a placeholder in backend/.env."
}

$namespaceResult = & kubectl get namespace $namespace --ignore-not-found -o name
if ($LASTEXITCODE -ne 0) {
    throw "Could not check whether namespace '$namespace' exists."
}
if (-not $namespaceResult) {
    & kubectl create namespace $namespace
    if ($LASTEXITCODE -ne 0) {
        throw "Could not create namespace '$namespace'."
    }
}

$existingSecretJson = (& kubectl get secret $secretName `
    --namespace $namespace `
    --ignore-not-found `
    --output=json | Out-String).Trim()
if ($LASTEXITCODE -ne 0) {
    throw "Could not check for the existing local Secret '$secretName'."
}

if ($existingSecretJson) {
    $existingSecret = $existingSecretJson | ConvertFrom-Json
    if (-not $existingSecret.data.POSTGRES_PASSWORD -or -not $existingSecret.data.REDIS_PASSWORD) {
        throw "The existing Secret '$secretName' is missing local database passwords; refusing to replace it."
    }
    $postgresPassword = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($existingSecret.data.POSTGRES_PASSWORD))
    $redisPassword = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($existingSecret.data.REDIS_PASSWORD))
} else {
    foreach ($claimName in @('postgres-data-livepulse-postgres-0', 'redis-data-livepulse-redis-0')) {
        $claimResult = & kubectl get pvc $claimName --namespace $namespace --ignore-not-found -o name
        if ($LASTEXITCODE -ne 0) {
            throw "Could not check for existing local data claim '$claimName'."
        }
        if ($claimResult) {
            throw "PVC '$claimName' already exists but '$secretName' does not. Refusing to generate new passwords for existing data."
        }
    }
    $postgresPassword = [System.Guid]::NewGuid().ToString('N')
    $redisPassword = [System.Guid]::NewGuid().ToString('N')
}

$secretLines = @(
    'POSTGRES_DB=livepulse'
    'POSTGRES_USER=livepulse'
    "POSTGRES_PASSWORD=$postgresPassword"
    "REDIS_PASSWORD=$redisPassword"
    "DATABASE_URL=postgres://livepulse:$postgresPassword@livepulse-postgres:5432/livepulse?sslmode=disable"
    "REDIS_URL=redis://default:$redisPassword@livepulse-redis:6379/0"
)

if (-not [string]::IsNullOrWhiteSpace($envValues['CLERK_SECRET_KEY'])) {
    $secretLines += "CLERK_SECRET_KEY=$($envValues['CLERK_SECRET_KEY'])"
}
if (-not [string]::IsNullOrWhiteSpace($envValues['ADMIN_API_KEY'])) {
    $secretLines += "ADMIN_API_KEY=$($envValues['ADMIN_API_KEY'])"
}
if ($EnableTicketmaster) {
    $secretLines += "EXTERNAL_API_KEY=$($envValues['EXTERNAL_API_KEY'])"
}

$tempEnvPath = [System.IO.Path]::GetTempFileName()
try {
    [System.IO.File]::WriteAllLines($tempEnvPath, $secretLines, [System.Text.UTF8Encoding]::new($false))

    $secretYaml = & kubectl create secret generic $secretName `
        --namespace $namespace `
        --from-env-file $tempEnvPath `
        --dry-run=client `
        --output=yaml
    if ($LASTEXITCODE -ne 0) {
        throw 'kubectl could not prepare the local Secret.'
    }

    $secretYaml | & kubectl apply -f -
    if ($LASTEXITCODE -ne 0) {
        throw 'kubectl could not apply the local Secret.'
    }

    if ($EnableTicketmaster) {
        Write-Host 'Ticketmaster key included. Starting the backend will issue its initial fetch and repeat every six hours.'
    } else {
        Write-Host 'Local PostgreSQL and Redis credentials created or preserved. Ticketmaster remains disabled.'
    }
} finally {
    Remove-Item -LiteralPath $tempEnvPath -Force -ErrorAction SilentlyContinue
}
