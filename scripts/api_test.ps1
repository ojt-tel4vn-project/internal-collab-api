$BASE = "http://localhost:8080/api/v1"
$PASS = 0
$FAIL = 0
$RESULTS = @()

function Invoke-RawAPI {
    param($Method, $Url, $Body, $Headers)
    try {
        $params = @{ Uri = $Url; Method = $Method; ErrorAction = 'Stop'; UseBasicParsing = $true }
        if ($Headers) { $params.Headers = $Headers }
        if ($Body) {
            $params.Body = ($Body | ConvertTo-Json -Depth 5)
            if (-not $params.Headers) { $params.Headers = @{} }
            $params.Headers['Content-Type'] = 'application/json'
        }
        $r = Invoke-WebRequest @params -SkipHttpErrorCheck 2>$null
        return @{ status = [int]$r.StatusCode; body = ($r.Content | ConvertFrom-Json -ErrorAction SilentlyContinue) }
    } catch {
        return @{ status = 0; body = $null }
    }
}

function Test-Endpoint {
    param($Name, $ExpectedStatus, $Method, $Url, $Body, $Headers)
    $r = Invoke-RawAPI $Method $Url $Body $Headers
    $ok = $r.status -eq $ExpectedStatus
    if ($ok) { $script:PASS++ } else { $script:FAIL++ }
    $icon = if ($ok) { "PASS" } else { "FAIL" }
    Write-Host "  [$icon] $Name | Got: $($r.status) (expected: $ExpectedStatus)"
    $script:RESULTS += [PSCustomObject]@{ Name=$Name; Expected=$ExpectedStatus; Got=$r.status; Pass=$ok }
    return $r
}

Write-Host ""
Write-Host "=========================================="
Write-Host " INTERNAL COLLAB API - FULL SYSTEM TEST  "
Write-Host "=========================================="

# =======================================================
Write-Host "`n[1] HEALTH"
Test-Endpoint "GET /health" 200 "GET" "http://localhost:8080/health" $null $null | Out-Null

# =======================================================
Write-Host "`n[2] AUTH"
$adminR = Invoke-RawAPI "POST" "$BASE/auth/login" @{email="admin@company.com";password="123456"} $null
$hrR    = Invoke-RawAPI "POST" "$BASE/auth/login" @{email="hr@company.com";password="123456"} $null
$mgR    = Invoke-RawAPI "POST" "$BASE/auth/login" @{email="manager@company.com";password="123456"} $null
$empR   = Invoke-RawAPI "POST" "$BASE/auth/login" @{email="staff@company.com";password="123456"} $null

foreach ($pair in @(
    @{n="Login Admin";r=$adminR}, @{n="Login HR";r=$hrR},
    @{n="Login Manager";r=$mgR}, @{n="Login Employee";r=$empR}
)) {
    $ok = $pair.r.status -eq 200
    if ($ok) { $PASS++ } else { $FAIL++ }
    Write-Host "  [$(if($ok){'PASS'}else{'FAIL'})] $($pair.n) | Got: $($pair.r.status) (expected: 200)"
    $RESULTS += [PSCustomObject]@{ Name=$pair.n; Expected=200; Got=$pair.r.status; Pass=$ok }
}

$badR = Invoke-RawAPI "POST" "$BASE/auth/login" @{email="no@x.com";password="bad"} $null
$ok = $badR.status -eq 401; if ($ok) { $PASS++ } else { $FAIL++ }
Write-Host "  [$(if($ok){'PASS'}else{'FAIL'})] Login wrong creds | Got: $($badR.status) (expected: 401)"
$RESULTS += [PSCustomObject]@{ Name="Login wrong creds -> 401"; Expected=401; Got=$badR.status; Pass=$ok }

$adminH = @{ Authorization = "Bearer $($adminR.body.token)" }
$hrH    = @{ Authorization = "Bearer $($hrR.body.token)" }
$mgH    = @{ Authorization = "Bearer $($mgR.body.token)" }
$empH   = @{ Authorization = "Bearer $($empR.body.token)" }

# =======================================================
Write-Host "`n[3] EMPLOYEE"
Test-Endpoint "GET /employees/me (employee)" 200 "GET" "$BASE/employees/me" $null $empH | Out-Null
Test-Endpoint "GET /employees (HR list)" 200 "GET" "$BASE/employees" $null $hrH | Out-Null
Test-Endpoint "GET /employees (no auth -> 401)" 401 "GET" "$BASE/employees" $null $null | Out-Null
Test-Endpoint "GET /employees/subordinates (manager)" 200 "GET" "$BASE/employees/subordinates" $null $mgH | Out-Null
Test-Endpoint "GET /employees/birthdays" 200 "GET" "$BASE/employees/birthdays" $null $hrH | Out-Null
Test-Endpoint "GET /employees/birthdays/config" 200 "GET" "$BASE/employees/birthdays/config" $null $hrH | Out-Null

# =======================================================
Write-Host "`n[4] LEAVE MANAGEMENT"
$ltR = Invoke-RawAPI "GET" "$BASE/leave-types" $null $empH
$ltId = $ltR.body.data[0].id
$ok = $ltR.status -eq 200; if ($ok) { $PASS++ } else { $FAIL++ }
Write-Host "  [$(if($ok){'PASS'}else{'FAIL'})] GET /leave-types | Got: $($ltR.status)"
$RESULTS += [PSCustomObject]@{ Name="GET /leave-types"; Expected=200; Got=$ltR.status; Pass=$ok }

Test-Endpoint "GET /leave-quotas" 200 "GET" "$BASE/leave-quotas" $null $empH | Out-Null
Test-Endpoint "GET /leave-requests (employee)" 200 "GET" "$BASE/leave-requests" $null $empH | Out-Null
Test-Endpoint "GET /leave-requests/pending-approval (manager)" 200 "GET" "$BASE/leave-requests/pending-approval" $null $mgH | Out-Null
Test-Endpoint "GET /leave-requests/overview (HR)" 200 "GET" "$BASE/leave-requests/overview" $null $hrH | Out-Null

if ($ltId) {
    $from = (Get-Date).AddDays(60).ToString("yyyy-MM-dd")
    $to   = (Get-Date).AddDays(61).ToString("yyyy-MM-dd")
    $lrBody = @{ leave_type_id=$ltId; from_date=$from; to_date=$to; reason="API test leave"; contact_during_leave="0901234567" }
    Test-Endpoint "POST /leave-requests (create)" 200 "POST" "$BASE/leave-requests" $lrBody $empH | Out-Null
}

# =======================================================
Write-Host "`n[5] DOCUMENTS"
Test-Endpoint "GET /documents/categories" 200 "GET" "$BASE/documents/categories" $null $empH | Out-Null
Test-Endpoint "GET /documents (employee)" 200 "GET" "$BASE/documents" $null $empH | Out-Null
Test-Endpoint "GET /documents (HR)" 200 "GET" "$BASE/documents" $null $hrH | Out-Null

# =======================================================
Write-Host "`n[6] NOTIFICATIONS"
Test-Endpoint "GET /notifications" 200 "GET" "$BASE/notifications" $null $empH | Out-Null
Test-Endpoint "GET /notifications/preferences" 200 "GET" "$BASE/notifications/preferences" $null $empH | Out-Null
Test-Endpoint "GET /notifications/types" 200 "GET" "$BASE/notifications/types" $null $empH | Out-Null
Test-Endpoint "POST /notifications/read-all" 200 "POST" "$BASE/notifications/read-all" $null $empH | Out-Null

# =======================================================
Write-Host "`n[7] AUDIT LOGS"
Test-Endpoint "GET /audit-logs (admin)" 200 "GET" "$BASE/audit-logs" $null $adminH | Out-Null
Test-Endpoint "GET /audit-logs/actions (admin)" 200 "GET" "$BASE/audit-logs/actions" $null $adminH | Out-Null
Test-Endpoint "GET /audit-logs (employee -> 403)" 403 "GET" "$BASE/audit-logs" $null $empH | Out-Null

# =======================================================
Write-Host "`n[8] ATTENDANCE"
Test-Endpoint "GET /attendances (HR - all)" 200 "GET" "$BASE/attendances" $null $hrH | Out-Null
Test-Endpoint "GET /attendances (employee - own)" 200 "GET" "$BASE/attendances" $null $empH | Out-Null
Test-Endpoint "GET /attendances/config (HR)" 200 "GET" "$BASE/attendances/config" $null $hrH | Out-Null
Test-Endpoint "GET /attendances/config (employee -> 403)" 403 "GET" "$BASE/attendances/config" $null $empH | Out-Null
Test-Endpoint "GET /attendances/summary (HR)" 200 "GET" "$BASE/attendances/summary" $null $hrH | Out-Null
Test-Endpoint "GET /attendances/summary (employee -> 403)" 403 "GET" "$BASE/attendances/summary" $null $empH | Out-Null

$attCfgBody = @{ confirmation_deadline_days=7; auto_confirm_enabled=$true; reminder_before_deadline_days=2 }
Test-Endpoint "PUT /attendances/config (HR)" 200 "PUT" "$BASE/attendances/config" $attCfgBody $hrH | Out-Null

# =======================================================
Write-Host "`n[9] RBAC / SECURITY"
Test-Endpoint "No token -> 401" 401 "GET" "$BASE/employees" $null $null | Out-Null
Test-Endpoint "Invalid JWT -> 401" 401 "GET" "$BASE/employees/me" $null @{Authorization="Bearer bad.token.here"} | Out-Null
Test-Endpoint "Employee -> HR endpoint -> 403" 403 "GET" "$BASE/attendances/config" $null $empH | Out-Null
Test-Endpoint "Employee -> Audit Logs -> 403" 403 "GET" "$BASE/audit-logs" $null $empH | Out-Null

# =======================================================
$total = $PASS + $FAIL
$pct = [math]::Round($PASS/$total*100, 1)

Write-Host ""
Write-Host "=========================================="
Write-Host " FINAL RESULTS"
Write-Host "=========================================="
Write-Host " Total Tests : $total"
Write-Host " Passed      : $PASS  ✅"
Write-Host " Failed      : $FAIL  ❌"
Write-Host " Score       : $pct%"
Write-Host ""

if ($FAIL -gt 0) {
    Write-Host "--- FAILED TESTS ---"
    $RESULTS | Where-Object { -not $_.Pass } | ForEach-Object {
        Write-Host "  FAIL: $($_.Name)"
        Write-Host "        Expected=$($_.Expected) | Got=$($_.Got)"
    }
}
Write-Host "=========================================="
