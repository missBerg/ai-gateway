# Adopter Logos

This directory contains logos of organizations that use Envoy AI Gateway in production.

## Adding Your Organization's Logo

We'd love to feature your organization! To add your logo:

### 1. Prepare Your Logo
- **Format**: SVG preferred (PNG also acceptable)
- **Size**: Optimal dimensions are 240x160px or similar 3:2 ratio
- **Background**: Transparent or white background works best
- **File naming**: Use lowercase with hyphens (e.g., `my-company.svg`)

### 2. Submit Your Logo
Choose one of these methods:

#### Method A: Pull Request (Recommended)
1. Fork the [ai-gateway repository](https://github.com/envoyproxy/ai-gateway)
2. Add your logo file to `site/static/img/adopters/`
3. Edit `site/src/components/Adopters/index.tsx` to add your organization:
   ```typescript
   {
     name: 'Your Company Name',
     logoUrl: '/img/adopters/your-company.svg',
     url: 'https://yourcompany.com', // Optional: link to your website
     description: 'Brief description of how you use Envoy AI Gateway', // Optional
   },
   ```
4. Submit a pull request with the title "Add [Company Name] to adopters"

#### Method B: GitHub Issue
1. Go to [GitHub Issues](https://github.com/envoyproxy/ai-gateway/issues/new)
2. Create a new issue with the title "Add [Company Name] to adopters"
3. Attach your logo file and provide the following information:
   - Company name
   - Website URL (optional)
   - Brief description of usage (optional)

### 3. Guidelines
- Logos should represent organizations actively using Envoy AI Gateway in production
- Please ensure you have permission to use the logo
- Logos should be family-friendly and professional
- We reserve the right to remove logos that don't meet our community standards

### 4. Questions?
If you have questions about adding your logo, please:
- Check our [Support the Project](/support) page for more ways to contribute
- Ask in our [GitHub Discussions](https://github.com/envoyproxy/ai-gateway/discussions)
- Join our [Slack community](https://envoyproxy.slack.com/archives/C07Q4N24VAA)

Thank you for supporting Envoy AI Gateway! 🎉
