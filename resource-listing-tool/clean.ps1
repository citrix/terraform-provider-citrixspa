#!/usr/bin/env pwsh

param(
    [Parameter(Position = 0)]
    [int]$Parallelism = 1
)

# Clean up generated files
Remove-Item -Force -ErrorAction SilentlyContinue .terraform.lock.hcl, imports.tf, management_summary.md, provider.tf, spa_resources.tf, terraform.tfstate, terraform.tfstate.backup

Remove-Item -Recurse -Force -ErrorAction SilentlyContinue .terraform

# Set up environment (parallelism defaults to 1, pass arg to override)
. ./setup-env.ps1 $Parallelism

# Set up terraform
& ./spa_manager.ps1 -Setup

& ./spa_manager.ps1 -List -DebugOutput

Write-Host ""
Write-Host "To set up terraform parallelism: . ./setup-env.ps1 $Parallelism"
