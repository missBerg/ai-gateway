# Adopters Data

This directory contains individual JSON files for each organization that supports Envoy AI Gateway.

## Adding Your Organization

To add your organization to the adopters list, create a new JSON file in this directory.

### File Naming Convention

Name your file using your organization's name in lowercase with hyphens instead of spaces:

- **Envoy AI Gateway** → `envoy-ai-gateway.json`
- **Acme Corporation** → `acme-corporation.json`
- **My Startup Inc** → `my-startup-inc.json`
- **123 Tech Solutions** → `123-tech-solutions.json`

### File Format

Create a JSON file with the following structure:

```json
{
  "name": "Your Organization Name",
  "logoUrl": "/img/adopters/your-organization.svg",
  "url": "https://yourcompany.com",
  "description": "Optional brief description"
}
```

### Required Fields

- **`name`**: Your organization's display name
- **`logoUrl`**: Path to your logo file in `/site/static/img/adopters/`

### Optional Fields

- **`url`**: Your organization's website (will make the logo clickable)
- **`description`**: Brief description (currently not displayed but reserved for future use)

### Logo Requirements

1. **Format**: SVG preferred (PNG also acceptable)
2. **Dimensions**: 240x160px or similar 3:2 ratio
3. **Background**: Transparent or white background
4. **File Location**: Place your logo in `/site/static/img/adopters/`
5. **File Naming**: Use the same naming convention as your JSON file

### Example

If your organization is "Acme Corporation":

**File**: `acme-corporation.json`
```json
{
  "name": "Acme Corporation",
  "logoUrl": "/img/adopters/acme-corporation.svg",
  "url": "https://acme.com",
  "description": "Leading provider of innovative solutions"
}
```

**Logo**: `/site/static/img/adopters/acme-corporation.svg`

### Submitting Your Addition

You can add your organization in two ways:

#### Option 1: GitHub Issue
1. Create a [GitHub issue](https://github.com/envoyproxy/ai-gateway/issues/new) with title "Add [Organization Name] to adopters"
2. Attach your logo file
3. Provide the JSON content in the issue description
4. We'll create the files for you!

#### Option 2: Pull Request
1. Fork the [ai-gateway repository](https://github.com/envoyproxy/ai-gateway)
2. Add your logo to `site/static/img/adopters/`
3. Create your JSON file in `site/src/data/adopters/`
4. Submit a pull request

**Note**: No need to manually update any index files! The system automatically detects and includes all JSON files in this directory.

### Display Order

Adopters are displayed alphabetically by organization name, so your position will be determined automatically.

### Questions?

If you have questions about adding your organization:
- Check our [Support the Project](https://aigateway.envoyproxy.io/support) page
- Ask in [GitHub Discussions](https://github.com/envoyproxy/ai-gateway/discussions)
- Join our [Slack community](https://envoyproxy.slack.com/archives/C07Q4N24VAA)

Thank you for supporting Envoy AI Gateway! 🎉
