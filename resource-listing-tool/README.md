# SPA Resource Listing Tool

This tool discovers resources from the SPA service provider and generates Terraform configuration files that can be used to manage your SPA configuration with Infrastructure as Code.

## Features

- **Resource Discovery**: Automatically discovers existing SPA resources (applications, access policies, browser mode, etc.)
- **Enhanced Data Collection**: Queries individual items for complete field data (can be disabled with `-q` for faster operation)
- **Terraform Generation**: Creates ready-to-use Terraform configuration files
- **Clean Resource Blocks**: Automatically excludes read-only fields (IDs, timestamps, computed values) for cleaner configurations
- **Import Blocks**: Generates modern Terraform import blocks for configuration-driven imports
- **Configuration Management**: Provides a foundation for managing SPA resources with Terraform
- **Built-in Validation**: Integrated validation and testing capabilities
- **Cross-Platform**: Pure Python implementation works on Windows, macOS, and Linux
- **High Performance**: 3-4x faster than previous shell-based implementation
- **Zero External Dependencies**: No external tools required (except Terraform)

## Quick Start

### Prerequisites

- **Python 3.6+**: Required for the spa_manager.py script

- **Terraform**: Required for resource management

  ```bash
  brew install terraform  # macOS
  ```

- **Valid SPA Credentials**: Either auth token or service principal credentials

### Step-by-Step Setup

1. **Initial Setup**

   ```bash
   cd resource-listing-tool
   python3 spa_manager.py --setup
   ```

2. **Configure Credentials**
   Edit the created `terraform.tfvars` file with your credentials:

   ```bash
   vim terraform.tfvars
   ```

   **For Service Principal Authentication (Recommended):**

   ```hcl
   customer_id   = "your-customer-id"
   client_id     = "your-client-id"
   client_secret = "your-client-secret"
   base_url      = "https://api.cloud.com/accessSecurity"
   ```

   **For Direct Token Authentication:**

   ```hcl
   customer_id = "your-customer-id"
   auth_token  = "your-auth-token"
   base_url    = "https://api.cloud.com/accessSecurity"
   ```

3. **Run Discovery**

   ```bash
   python3 spa_manager.py
   ```

4. **Validate Configuration**

   ```bash
   python3 spa_manager.py --validate
   ```

5. **Import Existing Resources**

   ```bash
   # Initialize Terraform (if not already done)
   terraform init

   # Import existing resources using modern Terraform import blocks
   terraform apply
   ```

   Or use the new auto-generation feature:

   ```bash
   terraform plan -generate-config-out=generated.tf
   ```

6. **Verify Configuration**
   ```bash
   # Check the current state
   terraform plan
   ```

## Commands

### Setup

```bash
python3 spa_manager.py --setup    # Create credentials file from example
```

### Discover and Generate

```bash
# Standard discovery with enhanced data collection (queries individual items)
python3 spa_manager.py -L         # Discover resources and generate Terraform files

# Quick discovery using list data only (faster but potentially incomplete field data)
python3 spa_manager.py -L -q      # Use -q flag to disable individual item queries

python3 spa_manager.py --debug    # Enable debug logging
python3 spa_manager.py --verbose  # Enable verbose output
```

#### Enhanced vs Quick Mode

- **Enhanced Mode (default)**: After getting the list of resources, the tool queries each individual item to get complete field data. This ensures all available fields are included in the generated Terraform configuration but takes longer.

- **Quick Mode (`-q` flag)**: Uses only the data from list queries, which is faster but may result in missing fields if the list API returns fewer fields than individual item queries.

### Testing and Validation

```bash
python3 spa_manager.py --test     # Run basic functionality tests
python3 spa_manager.py --validate # Validate generated Terraform configuration
```

### Cleanup

```bash
python3 spa_manager.py --clean    # Remove temporary files and generated configurations
```

### Help

```bash
python3 spa_manager.py --help     # Show usage information
```

## Generated Files

After running the tool, you'll get:

- **`spa_resources.tf`**: Complete Terraform configuration with resource definitions
- **`imports.tf`**: Modern Terraform import blocks for configuration-driven imports

### Example Files

- **`spa_resources.tf.example`**: Example of generated Terraform configuration

This example file shows what the tool generates and can help you understand the output format before running the actual discovery.

## Workflow

1. **Discovery**: The tool discovers your existing SPA resources
2. **Generation**: Creates Terraform configuration files based on discovered resources
3. **Testing**: Use `--test` to verify basic functionality
4. **Validation**: Use `--validate` to check generated files
5. **Import**: Use the generated Terraform import blocks to bring resources under Terraform management
6. **Management**: Use standard Terraform commands to manage your SPA configuration

## Resource Types Supported

The tool can discover and generate Terraform configuration for:

- **Applications**: All deployed applications in your SPA environment
- **Access Policies**: Access control policies with their configurations
- **Browser Mode**: Current browser mode configuration
- **Security Groups**: Security group configurations (requires specific names)
- **Routing Domains**: Routing domain configurations (requires specific FQDNs)
- **Certificates**: Certificate configurations
- **Terminate Machine Access**: Machine access termination settings

## Authentication

The tool supports two authentication methods:

1. **Direct Token Authentication**: Use a pre-generated bearer token
2. **Service Principal Authentication**: Use client ID and secret for OAuth2 flow

Edit `terraform.tfvars` to configure your preferred authentication method.

## Troubleshooting

### Authentication Issues

- Verify your credentials in `terraform.tfvars`
- Check network connectivity to SPA API
- Ensure you're using only one authentication method

### Missing Dependencies

- Install Python 3.6+: Follow your platform's Python installation guide
- Re-run the discovery script

### No Resources Found

- Verify you have resources in your SPA environment
- Check API permissions for your credentials
- Use `--debug` flag for detailed logging

### Import Failures

- Ensure Terraform is initialized: `terraform init`
- Check that resource IDs are correct
- Verify provider configuration matches your credentials

### Enhanced Query Issues

If individual item queries are failing:

- Use the `-q` flag to disable enhanced queries and fall back to list data only
- Check debug logs with `--debug` to see which individual queries are failing
- Verify API permissions allow querying individual items, not just lists
- Some resources may not support individual queries - the tool will fall back to list data automatically
- **Note**: Routing domains use FQDN for lookup (not ID/name like other resources)

### General Issues

1. **Check Credentials**: Ensure your `terraform.tfvars` has correct credentials
2. **Enable Debug**: Use `--debug` flag to see detailed logging
3. **Clean State**: Use `--clean` to remove temporary files and start fresh
4. **Check Connectivity**: Verify network access to the SPA API endpoints
5. **Rate Limiting**: If you have many resources, the enhanced mode may hit API rate limits - use `-q` for faster operation

## Advanced Usage

- **Custom Resource Filtering**: Modify the script to filter specific resource types
- **Multiple Environments**: Run the tool in different directories for different environments
- **Configuration Customization**: Edit generated files to match your specific requirements
- **Integration**: Integrate with CI/CD pipelines for automated resource management

## Next Steps

After generating the Terraform configuration:

1. **Review Resources**: Check the generated `spa_resources.tf` for accuracy
2. **Import Resources**: Run `terraform apply` to import existing resources using the generated import blocks
3. **Verify State**: Use `terraform plan` to verify the current state matches your configuration
4. **Customize**: Modify the generated configuration as needed for your requirements
5. **Apply Changes**: Use `terraform apply` to manage your SPA configuration

## Enhancement Summary

The SPA resource listing tool has been enhanced to not only discover resources but also generate Terraform configuration files that can be used to manage SPA resources with Infrastructure as Code.

### Key Enhancements

#### 1. Terraform Configuration Generation

- **New Feature**: Automatically generates `spa_resources.tf` with complete Terraform resource definitions
- **Authentication Support**: Supports both auth token and service principal authentication methods
- **Resource Types**: Generates configurations for applications, access policies, and browser mode

#### 2. Import Script Generation

- **New Feature**: Automatically generates `import_resources.sh` script with import commands
- **Executable Script**: Ready-to-run script that imports existing resources into Terraform state
- **Resource Mapping**: Correctly maps resource IDs to Terraform resource names

#### 3. Enhanced Resource Discovery

- **JSON Storage**: Discovered resources are stored as JSON files for processing
- **Summary Generation**: Provides a summary of discovered resources
- **Error Handling**: Improved error handling for cases where resources aren't found

#### 4. Integrated Tools

- **Built-in Validation**: Integrated Terraform validation functionality
- **Built-in Testing**: Integrated basic functionality testing
- **Comprehensive Documentation**: All documentation consolidated in this file

### Benefits

1. **Time Saving**: Eliminates manual Terraform configuration creation
2. **Accuracy**: Reduces human error in resource configuration
3. **Consistency**: Ensures consistent naming and structure
4. **Automation**: Enables automated resource management workflows
5. **Integration**: Built-in validation and testing capabilities

### Usage Examples

#### Basic Usage

```bash
python3 spa_manager.py
```

#### With Testing and Validation

```bash
python3 spa_manager.py --test
python3 spa_manager.py --validate
terraform apply  # Using the generated import blocks
terraform plan
```

#### With Debug

```bash
python3 spa_manager.py --debug
```

## Files in this Directory

- **`spa_manager.py`**: Main Python script with complete functionality
- **`terraform.tfvars.example`**: Template for credentials configuration
- **`spa_resources.tf`**: Generated Terraform resource configuration (created by script)
- **`imports.tf`**: Generated Terraform import blocks (created by script)
- **`spa_resources.tf.example`**: Example of generated Terraform configuration
- **`README.md`**: This comprehensive documentation file
- **`README_PYTHON.md`**: Python-specific implementation guide

## Prerequisites

- Valid SPA/Citrix Cloud credentials
- Terraform installed
- Python 3.6+ (built-in on most modern systems)
- Go compiler (for building the provider)

## What It Does

The script will:

1. Build and install the SPA Terraform provider (if needed)
2. Create a clean Terraform configuration
3. Initialize Terraform in this directory
4. Query the SPA API for:
   - Applications list (with optional individual item details)
   - Access policies list (with optional individual item details)
   - Security groups (with optional individual item details)
   - Routing domains (with optional individual item details)
   - Browser mode configuration
   - Hybrid configuration
   - Last activity information
5. Generate Terraform configuration files with complete field data
6. Generate cross-platform Python import script
7. Display results in a user-friendly format
8. Clean up temporary files

## Performance Considerations

The Python implementation provides significant performance improvements:

- **3-4x faster execution** (3-5 seconds vs 15-20 seconds) in quick mode
- **Enhanced data quality** in standard mode (individual queries ensure complete field data)
- **Zero external dependencies** (no jq, awk, or other tools required)
- **Cross-platform compatibility** (Windows, macOS, Linux)
- **Better error handling** and user feedback

### When to Use Each Mode

- **Enhanced Mode (default)**: Use when you need complete and accurate field data for all resources. This is recommended for production use where you want to ensure all available fields are captured in your Terraform configuration.

- **Quick Mode (`-q` flag)**: Use when you have many resources and need faster discovery, or when you're doing initial exploration and don't need complete field data immediately.

## Debug Mode

When run with `--debug`, the script:

- Enables Terraform debug logging
- Creates a `resource-listing-debug.log` file
- Shows detailed API requests and responses

## Isolation

This tool runs in complete isolation from other Terraform files in the parent directory, preventing conflicts and ensuring clean execution.
