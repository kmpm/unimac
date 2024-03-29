
Set-DotEnv

$date = Get-Date -Format "yyyy-MM-dd_HHmm"


Write-Host "clients"
& '.\unimac.exe'  clients -output out\today\clients.xlsx
Copy-Item out\today\clients.xlsx -Destination "out\results\$($date)_clients.xlsx"


Write-Host "devices"
& '.\unimac.exe'  devices -output out\today\devices.xlsx
Copy-Item out\today\devices.xlsx -Destination "out\results\$($date)_devices.xlsx"
