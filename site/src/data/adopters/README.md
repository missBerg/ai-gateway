# Adopters Data

This directory contains the adopters data for organizations that support Envoy AI Gateway.

Each adopter is stored in their own JSON file to avoid merge conflicts and make contributions easier!

## Adding Your Organization

We've made it super easy to add your organization! You can either use our issue form (recommended) or create a file directly.

### Option 1: Issue Form (Recommended for New Contributors)

**[Submit an Adopter Form →](https://github.com/envoyproxy/ai-gateway/issues/new?template=add-adopter.yml)**

Fill out the form and follow the instructions. You'll be guided through creating your adopter file.

### Option 2: Direct PR (For GitHub Users)

1. **Fork this repository**
2. **Copy the template**: Duplicate `_template.json` and rename it to `your-company-name.json`
3. **Edit your file**: Update the JSON with your organization's details
4. **Submit a PR**: Create a pull request with your new file

No merge conflicts since everyone adds their own file! 🎉

### File Naming

- Use lowercase with hyphens: `my-company.json`
- Be descriptive but concise
- Don't use special characters or spaces

**Examples:**
- ✅ `acme-corp.json`
- ✅ `awesome-startup.json`
- ❌ `My Company!.json`
- ❌ `company123.json`

### JSON Format

Each adopter file should contain:

```json
{
  "name": "Your Organization Name",
  "logoUrl": "https://yoursite.com/logo.svg",
  "url": "https://yourcompany.com",
  "description": "Brief description (shown on hover)"
}
```

### Fields

- **`name`** (required): Your organization's display name
- **`logoUrl`** (required): URL to your logo - can be either:
  - External URL: `https://yoursite.com/logo.svg` (easiest!)
  - Local path: `/img/adopters/your-logo.svg` (requires uploading logo file)
- **`url`** (optional): Your organization's website (makes the logo clickable)
- **`description`** (optional): Brief description shown when users hover over your logo

### Logo Options

#### Option 1: External URL (Easiest)
Simply provide a direct link to your logo hosted anywhere:
```json
"logoUrl": "https://yourcompany.com/assets/logo.svg"
```

#### Option 2: Local Logo (Better Performance)
If you prefer to host the logo locally:
1. Add your logo to `site/static/img/adopters/your-company.svg`
2. Reference it as: `"logoUrl": "/img/adopters/your-company.svg"`

**Logo Specifications:**
- **Format**: SVG preferred (PNG also acceptable)
- **Dimensions**: 240x160px or similar 3:2 ratio recommended
- **Background**: Transparent or white background works best

### Example

Create a file named `acme-corp.json`:

```json
{
  "name": "Acme Corporation",
  "logoUrl": "https://acme.com/logo.svg",
  "url": "https://acme.com",
  "description": "Using Envoy AI Gateway for multi-model AI routing"
}
```

### Display Order

Adopters are displayed alphabetically by organization name, so your position will be determined automatically.

### Need Help?

If you have questions about adding your organization:
- Check our [Support the Project](https://aigateway.envoyproxy.io/support) page
- Ask in [GitHub Discussions](https://github.com/envoyproxy/ai-gateway/discussions)
- Join our [Slack community](https://envoyproxy.slack.com/archives/C07Q4N24VAA)

Thank you for supporting Envoy AI Gateway!
