#!/usr/bin/env pwsh
# Configure Terraform parallelism: . ./setup-env.ps1 [parallelism]

param([int]$Parallelism = 1)

$env:TF_CLI_ARGS_apply = "-parallelism=$Parallelism"
$env:TF_CLI_ARGS_plan = "-parallelism=$Parallelism"

Write-Host "✓ Terraform parallelism set to $Parallelism"
