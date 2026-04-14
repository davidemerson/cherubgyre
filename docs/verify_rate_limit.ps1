$username = "rate_limit_user_" + (Get-Random)
$masterCode = "4f88690e-0fbc-47b9-88e3-2d5ee2ac03d2"
$baseUrl = "http://localhost:8080"

Write-Host "Registering user: $username"
$registerBody = @{
    username = $username
    normal_pin = "1234"
    duress_pin = "9999"
    invite_code = $masterCode
} | ConvertTo-Json

$registerResponse = Invoke-RestMethod -Uri "$baseUrl/register" -Method Post -Body $registerBody -ContentType "application/json"
Write-Host "Registered."

Write-Host "Logging in..."
$loginBody = @{
    username = $username
    normal_pin = "1234"
} | ConvertTo-Json

$loginResponse = Invoke-RestMethod -Uri "$baseUrl/login" -Method Post -Body $loginBody -ContentType "application/json"
$token = $loginResponse.token
Write-Host "Logged in. Token obtained."

$headers = @{
    Authorization = $token
}

Write-Host "Testing Rate Limit (Limit: 3/hour)..."

for ($i = 1; $i -le 4; $i++) {
    Write-Host "Attempt $i :" -NoNewline
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/invite" -Method Get -Headers $headers
        Write-Host " Success - Code: $($response.inviteCode)"
    } catch {
        Write-Host " Failed - $($_.Exception.Message)"
        # Print detailed error if available
        if ($_.Exception.Response) {
             $reader = New-Object System.IO.StreamReader $_.Exception.Response.GetResponseStream()
             $errBody = $reader.ReadToEnd()
             Write-Host "   Error Body: $errBody"
        }
    }
}
