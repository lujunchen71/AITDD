# AITDD JSON File Selection Menu
# Use arrow keys to select, Enter to confirm, ESC to exit

param(
    [string]$DebugDir
)

# Check if directory exists
if (-not (Test-Path $DebugDir)) {
    Write-Host "[ERROR] Directory not found: $DebugDir" -ForegroundColor Red
    exit 1
}

# Get all JSON files
$jsonFiles = Get-ChildItem -Path $DebugDir -Filter "*.json" | Sort-Object Name

if ($jsonFiles.Count -eq 0) {
    Write-Host "[ERROR] No .json files found in directory" -ForegroundColor Red
    exit 1
}

$selectedIndex = 0
$maxIndex = $jsonFiles.Count - 1
$selectionMade = $false

function Draw-Menu {
    # Clear only the menu area
    $cursorTop = [Console]::CursorTop
    
    # Move cursor to top
    [Console]::SetCursorPosition(0, 0)
    [Console]::Clear()
    
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "   AITDD Database Import Tool" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Select a JSON file to import:" -ForegroundColor Yellow
    Write-Host ""
    
    for ($i = 0; $i -lt $jsonFiles.Count; $i++) {
        if ($i -eq $selectedIndex) {
            # Highlight selected item
            Write-Host "  > " -NoNewline -ForegroundColor Green
            $size = [math]::Round($jsonFiles[$i].Length / 1KB, 2)
            Write-Host "$($jsonFiles[$i].Name) ($size KB)" -ForegroundColor Black -BackgroundColor Green
        } else {
            Write-Host "    " -NoNewline
            $size = [math]::Round($jsonFiles[$i].Length / 1KB, 2)
            Write-Host "$($jsonFiles[$i].Name) ($size KB)" -ForegroundColor White
        }
    }
    
    Write-Host ""
    Write-Host "========================================" -ForegroundColor DarkGray
    Write-Host "Keys: [Up/Down] Select | [Enter] Confirm | [ESC] Cancel" -ForegroundColor DarkGray
    Write-Host ""
}

# Initial draw
Draw-Menu

# Main loop
while (-not $selectionMade) {
    $key = $Host.UI.RawUI.ReadKey('NoEcho,IncludeKeyDown')
    
    switch ($key.VirtualKeyCode) {
        38 { # Up arrow
            if ($selectedIndex -gt 0) { 
                $selectedIndex-- 
                Draw-Menu
            }
        }
        40 { # Down arrow
            if ($selectedIndex -lt $maxIndex) { 
                $selectedIndex++ 
                Draw-Menu
            }
        }
        13 { # Enter
            $selectionMade = $true
            # Clear and show selection
            [Console]::Clear()
            Write-Host "========================================" -ForegroundColor Cyan
            Write-Host "   AITDD Database Import Tool" -ForegroundColor Cyan
            Write-Host "========================================" -ForegroundColor Cyan
            Write-Host ""
            Write-Host "Selected: " -NoNewline
            $size = [math]::Round($jsonFiles[$selectedIndex].Length / 1KB, 2)
            Write-Host "$($jsonFiles[$selectedIndex].Name) ($size KB)" -ForegroundColor Green
            Write-Host ""
            # Output selected file path for bat script
            Write-Output $jsonFiles[$selectedIndex].FullName
        }
        27 { # ESC
            $selectionMade = $true
            [Console]::Clear()
            Write-Host "Operation cancelled." -ForegroundColor Yellow
        }
    }
}
