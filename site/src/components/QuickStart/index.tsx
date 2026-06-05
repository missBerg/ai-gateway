import React, {useState} from 'react';
import Link from '@docusaurus/Link';
import Heading from '@theme/Heading';
import CodeBlock from '@theme/CodeBlock';
import vars from '@site/docs/_vars.json';
import styles from './styles.module.css';

// Versions come from the same single source the docs use (docs/_vars.json), so
// the landing commands stay in lockstep with installation.md / prerequisites.md
// and the Envoy Gateway + AI Gateway versions remain mutually compatible.
const AIGW = `v${vars.aigwVersion}`;
const EG = `v${vars.egVersion}`;
const REF = vars.aigwGitRef;

const STEP_EG = `helm upgrade -i eg oci://docker.io/envoyproxy/gateway-helm \\
    --version ${EG} \\
    --namespace envoy-gateway-system --create-namespace \\
    -f https://raw.githubusercontent.com/envoyproxy/ai-gateway/${REF}/manifests/envoy-gateway-values.yaml

kubectl wait --timeout=2m -n envoy-gateway-system \\
    deployment/envoy-gateway --for=condition=Available`;

const STEP_CRDS = `helm upgrade -i aieg-crd oci://docker.io/envoyproxy/ai-gateway-crds-helm \\
    --version ${AIGW} \\
    --namespace envoy-ai-gateway-system --create-namespace`;

const STEP_CTRL = `helm upgrade -i aieg oci://docker.io/envoyproxy/ai-gateway-helm \\
    --version ${AIGW} \\
    --namespace envoy-ai-gateway-system --create-namespace

kubectl wait --timeout=2m -n envoy-ai-gateway-system \\
    deployment/ai-gateway-controller --for=condition=Available`;

const STEPS = [
  {
    n: 1,
    tab: 'Envoy Gateway',
    title: 'Install Envoy Gateway',
    desc: 'With the AI Gateway values file applied.',
    code: STEP_EG,
  },
  {
    n: 2,
    tab: 'AI Gateway CRDs',
    title: 'Install the AI Gateway CRDs',
    desc: 'The custom resource definitions.',
    code: STEP_CRDS,
  },
  {
    n: 3,
    tab: 'Controller',
    title: 'Install the AI Gateway controller',
    desc: 'Then wait for it to become available.',
    code: STEP_CTRL,
  },
];

export default function QuickStart(): React.ReactElement {
  const [active, setActive] = useState(0);
  const step = STEPS[active];

  return (
    <section className={styles.section} aria-labelledby="quickstart-heading">
      <div className="container">
        <div className={styles.split}>
          {/* Left column: heading + blurb */}
          <div className={styles.intro}>
            <span className="sectionEyebrow sectionEyebrow--purple">Quick start</span>
            <Heading as="h2" id="quickstart-heading" className={styles.introTitle}>
              Up and running in three steps
            </Heading>
            <p className={styles.introText}>
              Install onto any Kubernetes {vars.k8sMinVersion}+ cluster with Helm. Need a cluster or
              more detail? <Link to="/docs/getting-started/">Read the full getting-started guide →</Link>
            </p>
            <p className={styles.note}>
              You&apos;re viewing main-branch docs, so these track the latest build. Pin to a release
              before production — see the{' '}
              <Link to="/docs/getting-started/installation">installation guide</Link>.
            </p>
          </div>

          {/* Right column: tabbed install panel */}
          <div className={styles.panel}>
            <div className={styles.tabs} role="tablist" aria-label="Installation steps">
              {STEPS.map((s, i) => (
                <button
                  key={s.n}
                  type="button"
                  role="tab"
                  aria-selected={i === active}
                  className={styles.tab}
                  data-active={i === active}
                  onClick={() => setActive(i)}>
                  <span className={styles.tabNum} aria-hidden="true">{s.n}</span>
                  <span className={styles.tabLabel}>{s.tab}</span>
                </button>
              ))}
            </div>

            <div className={styles.panelBody} role="tabpanel">
              <div className={styles.panelMeta}>
                <Heading as="h3" className={styles.stepTitle}>{step.title}</Heading>
                <p className={styles.stepDesc}>{step.desc}</p>
              </div>
              <CodeBlock language="bash" className={styles.code}>
                {step.code}
              </CodeBlock>
            </div>

            <div className={styles.panelFoot}>
              <span className={styles.prereqLabel}>Prerequisites</span>
              <span className={styles.chip}>Kubernetes {vars.k8sMinVersion}+</span>
              <span className={styles.chip}>kubectl</span>
              <span className={styles.chip}>helm</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
