import React from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import Heading from '@theme/Heading';
import { Star, MessageCircle, Bug, Wrench, Building, FileText, Heart, Rocket, BookOpen, Zap, Plug, ExternalLink, Github } from 'lucide-react';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@site/src/components/ui/accordion';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@site/src/components/ui/card';
import styles from './support.module.css';

const supportWays = [
  {
    icon: Star,
    title: 'Star & Follow on GitHub',
    description: 'Show your support and help us understand community interest',
    action: {
      text: 'Star on GitHub',
      href: 'https://github.com/envoyproxy/ai-gateway',
      primary: true
    },
    highlight: true
  },
  {
    icon: Building,
    title: 'Add Your Adopter Logo',
    description: 'Show your support by adding your organization\'s logo to the adopters list',
    action: {
      text: 'Add Your Logo',
      href: '#add-your-logo',
      primary: true
    },
    highlight: true
  },
  {
    icon: MessageCircle,
    title: 'Join Our Community',
    description: 'Connect with users, contributors, and maintainers',
    items: [
      { text: 'Join Slack', href: 'https://envoyproxy.slack.com/archives/C07Q4N24VAA' },
      { text: 'Weekly Meetings', href: 'https://zoom-lfx.platform.linuxfoundation.org/meeting/91546415944?password=61fd5a5d-41e9-4b0c-86ea-b607c4513e37' },
      { text: 'Meeting Notes', href: 'https://docs.google.com/document/d/10e1sfsF-3G3Du5nBHGmLjXw5GVMqqCvFDqp_O65B0_w/edit?tab=t.0' }
    ]
  },
  {
    icon: Bug,
    title: 'Report Issues & Features',
    description: 'Help us improve by reporting bugs and suggesting features',
    items: [
      { text: 'Report Bug', href: 'https://github.com/envoyproxy/ai-gateway/issues/new?assignees=&labels=bug&projects=&template=bug_report.yml' },
      { text: 'Request Feature', href: 'https://github.com/envoyproxy/ai-gateway/issues/new?assignees=&labels=enhancement&projects=&template=feature_request.yml' },
      { text: 'Ask Question', href: 'https://github.com/envoyproxy/ai-gateway/discussions/new?category=q-a' }
    ]
  },
  {
    icon: Wrench,
    title: 'Contribute Code',
    description: 'Ready to dive into the code? We welcome contributions of all sizes',
    action: {
      text: 'View Contributing Guide',
      href: 'https://github.com/envoyproxy/ai-gateway/blob/main/CONTRIBUTING.md',
      primary: false
    }
  },
  {
    icon: FileText,
    title: 'Improve Documentation',
    description: 'Help make Envoy AI Gateway more accessible',
    action: {
      text: 'Edit Documentation',
      href: 'https://github.com/envoyproxy/ai-gateway/tree/main/site/docs',
      primary: false
    }
  }
];

const codeContributionAreas = [
  { icon: BookOpen, text: 'Documentation improvements' },
  { icon: Bug, text: 'Bug fixes and testing' },
  { icon: Plug, text: 'New LLM provider integrations' },
  { icon: Zap, text: 'Performance optimizations' },
  { icon: FileText, text: 'Example applications and tutorials' }
];

function SupportCard({ icon: IconComponent, title, description, action, items, highlight }: any) {
  return (
    <Card className={`${styles.modernCard} ${highlight ? styles.highlight : ''}`}>
      <CardHeader className={styles.cardHeader}>
        <div className={styles.cardIconWrapper}>
          <IconComponent size={48} className={styles.cardIcon} />
        </div>
        <div>
          <CardTitle className={styles.cardTitle}>{title}</CardTitle>
          <CardDescription className={styles.cardDescription}>{description}</CardDescription>
        </div>
      </CardHeader>

      <CardContent className={styles.cardContentSection}>
        {items && (
          <div className={`${styles.cardLinks} button`}>
            {items.map((item: any, idx: number) => (
              <Link key={idx} href={item.href} className={styles.cardLink}>
                {item.text}
              </Link>
            ))}
          </div>
        )}
      </CardContent>

      {action && (
        <CardFooter className={styles.cardFooter}>
          <Link
            className={`button ${action.primary ? 'button--primary' : 'button--secondary'} ${styles.cardButton}`}
            href={action.href}>
            {action.text}
          </Link>
        </CardFooter>
      )}
    </Card>
  );
}

function HeroSection() {
  return (
    <section className={styles.hero}>
      <div className="container">
        <div className={styles.heroContent}>
          <Heading as="h1" className={styles.heroTitle}>
            Support the Project
          </Heading>
          <p className={styles.heroSubtitle}>
            Envoy AI Gateway thrives because of our amazing community! There are many ways you can help keep the project alive and growing. Every contribution, no matter how small, makes a difference.
          </p>
          <div className={styles.heroButtons}>
            <Link
              className="button button--primary button--lg"
              href="https://github.com/envoyproxy/ai-gateway">
              <Star size={20} />
              Star on GitHub
            </Link>
            <Link
              className="button button--secondary button--lg"
              href="https://envoyproxy.slack.com/archives/C07Q4N24VAA">
              <MessageCircle size={20} />
              Join Slack
            </Link>
          </div>
        </div>
      </div>
    </section>
  );
}

function AdopterSection() {
  return (
    <section className={styles.adopterSection} id="add-your-logo">
      <div className="container">
        <div className={styles.sectionHeader}>
          <Heading as="h2" className={styles.sectionTitle}>
            <Building size={36} />
            How to Add Your Organization as an Adopter
          </Heading>
          <p className={styles.sectionDescription}>
            Show your support for Envoy AI Gateway! Adding your organization's logo helps demonstrate community backing and encourages others to try the project. Adopters are displayed alphabetically by organization name.
          </p>
        </div>

        <div className={styles.modernAdopterGuide}>
          <div className={styles.primaryMethod}>
            <div className={styles.methodCard}>
              <div className={styles.methodIcon}>
                <Github size={48} />
              </div>
              <h3>Quick & Easy: Edit on GitHub</h3>
              <p>
                Click the button below to edit the adopters file directly in your browser.
                Add your organization's details, and GitHub will automatically create a pull request for you!
              </p>
              <div className={styles.actionButtons}>
                <Link
                  className="button button--primary button--lg"
                  href="https://github.com/envoyproxy/ai-gateway/edit/main/site/src/data/adopters/adopters.json"
                  target="_blank"
                  rel="noopener noreferrer">
                  <Github size={20} />
                  Edit adopters.json on GitHub
                </Link>
              </div>
            </div>
          </div>

          <Accordion type="single" collapsible className={styles.accordion}>
            <AccordionItem value="how-to-add" className={styles.accordionItem}>
              <AccordionTrigger className={styles.accordionTrigger}>
                <div className={styles.optionHeader}>
                  <div className={styles.optionIcon}>
                    <FileText size={24} />
                  </div>
                  <div className={styles.optionInfo}>
                    <h3>What to Add</h3>
                    <p>JSON format and field descriptions</p>
                  </div>
                </div>
              </AccordionTrigger>
              <AccordionContent className={styles.accordionContent}>
                <div className={styles.requirementsList}>
                  <div className={styles.requirementGroup}>
                    <h4>Add Your Entry</h4>
                    <p>Add this JSON object to the array in <code>adopters/adopters.json</code>:</p>
                    <div className={styles.jsonExample}>
                      <pre>{`{
  "name": "Your Organization Name",
  "logoUrl": "https://yoursite.com/logo.svg",
  "url": "https://yourcompany.com",
  "description": "Brief description (shown on hover)"
}`}</pre>
                    </div>
                  </div>
                  <div className={styles.requirementGroup}>
                    <h4>Fields</h4>
                    <ul>
                      <li><strong>name</strong> (required): Your organization's display name</li>
                      <li><strong>logoUrl</strong> (required): URL to your logo - can be external (https://...) or local (/img/adopters/...)</li>
                      <li><strong>url</strong> (optional): Your organization's website</li>
                      <li><strong>description</strong> (optional): Brief description shown on hover</li>
                    </ul>
                  </div>
                  <div className={styles.requirementGroup}>
                    <h4>Logo Options</h4>
                    <p><strong>Option 1 (Easiest):</strong> Use an external URL to your logo</p>
                    <p><code>"logoUrl": "https://yourcompany.com/logo.svg"</code></p>
                    <p><strong>Option 2:</strong> Upload logo to <code>site/static/img/adopters/</code> and reference it</p>
                    <p><code>"logoUrl": "/img/adopters/your-logo.svg"</code></p>
                    <p><em>Logo specs: SVG preferred, 240x160px or 3:2 ratio, transparent/white background</em></p>
                  </div>
                </div>
              </AccordionContent>
            </AccordionItem>

            <AccordionItem value="alternative" className={styles.accordionItem}>
              <AccordionTrigger className={styles.accordionTrigger}>
                <div className={styles.optionHeader}>
                  <div className={styles.optionIcon}>
                    <Wrench size={24} />
                  </div>
                  <div className={styles.optionInfo}>
                    <h3>Alternative: GitHub Issue</h3>
                    <p>Let us add it for you</p>
                  </div>
                </div>
              </AccordionTrigger>
              <AccordionContent className={styles.accordionContent}>
                <div className={styles.stepByStep}>
                  <div className={styles.step}>
                    <div className={styles.stepNumber}>1</div>
                    <div className={styles.stepContent}>
                      <h4>Create GitHub Issue</h4>
                      <p>Create a new issue with title "Add [Organization Name] to adopters"</p>
                    </div>
                  </div>
                  <div className={styles.step}>
                    <div className={styles.stepNumber}>2</div>
                    <div className={styles.stepContent}>
                      <h4>Provide Details</h4>
                      <p>Include your organization name, logo URL or file, website URL (optional), and description (optional)</p>
                    </div>
                  </div>
                  <div className={styles.actionButtons}>
                    <Link
                      className="button button--primary"
                      href="https://github.com/envoyproxy/ai-gateway/issues/new">
                      <Github size={16} />
                      Create Issue
                    </Link>
                  </div>
                </div>
              </AccordionContent>
            </AccordionItem>
          </Accordion>
        </div>
      </div>
    </section>
  );
}

function ThankYouSection() {
  return (
    <section className={styles.thankYou}>
      <div className="container">
        <div className={styles.thankYouContent}>
          <h2>
            <Heart size={36} />
            Thank You!
          </h2>
          <p>
            Every form of support matters, whether you're starring the repo, reporting a bug, or contributing code.
            The Envoy AI Gateway project is what it is today because of contributors like you.
          </p>
          <p className={styles.tagline}>
            <strong>Together, we're building the future of AI traffic management!</strong>
            <Rocket size={24} />
          </p>
          <div className={styles.communityLinks}>
            <Link href="https://envoyproxy.slack.com/archives/C07Q4N24VAA">
              Join our Slack community
            </Link>
            <span>•</span>
            <Link href="https://github.com/envoyproxy/ai-gateway/discussions">
              Start a discussion on GitHub
            </Link>
          </div>
        </div>
      </div>
    </section>
  );
}

export default function Support(): React.ReactElement {
  return (
    <Layout
      title="Support the Project"
      description="Learn how you can support and contribute to Envoy AI Gateway">
      <HeroSection />

      <section className={styles.supportGrid}>
        <div className="container">
          <div className={styles.grid}>
            {supportWays.map((way, idx) => (
              <SupportCard key={idx} {...way} />
            ))}
          </div>
        </div>
      </section>

      <AdopterSection />
      <ThankYouSection />
    </Layout>
  );
}
