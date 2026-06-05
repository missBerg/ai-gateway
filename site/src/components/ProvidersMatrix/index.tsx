import React, {useCallback, useMemo, useState} from 'react';
import Link from '@docusaurus/Link';
import {
  AUTH_META,
  PROVIDERS,
  type Provider,
  type ProviderStatus,
  type SchemaVariant,
} from '@site/src/data/providers';
import styles from './styles.module.css';

/* ---------- Inline SVG icons (no icon-library dependency) ---------- */
type IconProps = {size?: number; strokeWidth?: number; className?: string};

function Svg({
  size = 16,
  strokeWidth = 2,
  className,
  children,
}: IconProps & {children: React.ReactNode}): React.ReactElement {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={strokeWidth}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      aria-hidden="true">
      {children}
    </svg>
  );
}

const CheckCircleIcon = (p: IconProps) => (
  <Svg {...p}>
    <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
    <polyline points="22 4 12 14.01 9 11.01" />
  </Svg>
);
const AlertTriangleIcon = (p: IconProps) => (
  <Svg {...p}>
    <path d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
    <line x1="12" y1="9" x2="12" y2="13" />
    <line x1="12" y1="17" x2="12.01" y2="17" />
  </Svg>
);
const ClockIcon = (p: IconProps) => (
  <Svg {...p}>
    <circle cx="12" cy="12" r="10" />
    <polyline points="12 6 12 12 16 14" />
  </Svg>
);
const ChevronUpIcon = (p: IconProps) => (
  <Svg {...p}>
    <polyline points="18 15 12 9 6 15" />
  </Svg>
);
const ChevronDownIcon = (p: IconProps) => (
  <Svg {...p}>
    <polyline points="6 9 12 15 18 9" />
  </Svg>
);
const ChevronsUpDownIcon = (p: IconProps) => (
  <Svg {...p}>
    <path d="m7 15 5 5 5-5" />
    <path d="m7 9 5-5 5 5" />
  </Svg>
);
const CopyIcon = (p: IconProps) => (
  <Svg {...p}>
    <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
  </Svg>
);
const CheckIcon = (p: IconProps) => (
  <Svg {...p}>
    <polyline points="20 6 9 17 4 12" />
  </Svg>
);
const SearchIcon = (p: IconProps) => (
  <Svg {...p}>
    <circle cx="11" cy="11" r="8" />
    <line x1="21" y1="21" x2="16.65" y2="16.65" />
  </Svg>
);
const ExternalLinkIcon = (p: IconProps) => (
  <Svg {...p}>
    <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
    <polyline points="15 3 21 3 21 9" />
    <line x1="10" y1="14" x2="21" y2="3" />
  </Svg>
);

type SortKey = 'name' | 'status';
type SortDir = 'asc' | 'desc';

// Status ordering — supported first, then partial, then planned.
const STATUS_RANK: Record<ProviderStatus, number> = {
  supported: 0,
  partial: 1,
  planned: 2,
};

const STATUS_META: Record<
  ProviderStatus,
  {Icon: (p: IconProps) => React.ReactElement; label: string; cls: string}
> = {
  supported: {Icon: CheckCircleIcon, label: 'Supported', cls: styles.statusOk},
  partial: {Icon: AlertTriangleIcon, label: 'Partial', cls: styles.statusWarn},
  planned: {Icon: ClockIcon, label: 'Planned', cls: styles.statusPlanned},
};

function StatusCell({status}: {status: ProviderStatus}): React.ReactElement {
  const {Icon, label, cls} = STATUS_META[status];
  return (
    <span className={`${styles.status} ${cls}`} title={label}>
      <Icon size={16} />
      <span className={styles.statusLabel}>{label}</span>
    </span>
  );
}

function formatConfig(variant: SchemaVariant): string {
  // Pretty-print for display and copy; two-space indent matches docs YAML/JSON.
  return JSON.stringify(variant.config, null, 2);
}

function SchemaBlock({variant}: {variant: SchemaVariant}): React.ReactElement {
  const [copied, setCopied] = useState(false);
  const text = useMemo(() => formatConfig(variant), [variant]);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 1600);
    } catch {
      /* clipboard unavailable */
    }
  }, [text]);

  return (
    <div className={styles.schemaBlock}>
      {variant.label && <span className={styles.schemaLabel}>{variant.label}</span>}
      <div className={styles.codeWrap}>
        <button
          type="button"
          className={styles.copyBtn}
          onClick={handleCopy}
          aria-label={copied ? 'Copied' : 'Copy schema config'}>
          {copied ? <CheckIcon size={13} strokeWidth={2.4} /> : <CopyIcon size={13} />}
        </button>
        <pre className={styles.codeBlock}>
          <code>{text}</code>
        </pre>
      </div>
    </div>
  );
}

function AuthList({auth}: {auth: Provider['auth']}): React.ReactElement {
  return (
    <ul className={styles.authList}>
      {auth.map((kind) => {
        const meta = AUTH_META[kind];
        return (
          <li key={kind}>
            {meta.href ? (
              <Link to={meta.href} className={styles.authLink}>
                {meta.label}
              </Link>
            ) : (
              <span className={styles.authMuted}>{meta.label}</span>
            )}
          </li>
        );
      })}
    </ul>
  );
}

function SortIcon({active, dir}: {active: boolean; dir: SortDir}): React.ReactElement {
  if (!active) {
    return <ChevronsUpDownIcon size={13} className={styles.sortIcon} />;
  }
  return dir === 'asc' ? (
    <ChevronUpIcon size={13} strokeWidth={2.2} className={`${styles.sortIcon} ${styles.sortIconActive}`} />
  ) : (
    <ChevronDownIcon size={13} strokeWidth={2.2} className={`${styles.sortIcon} ${styles.sortIconActive}`} />
  );
}

export default function ProvidersMatrix(): React.ReactElement {
  const [sortKey, setSortKey] = useState<SortKey>('name');
  const [sortDir, setSortDir] = useState<SortDir>('asc');
  const [query, setQuery] = useState('');

  const toggleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortDir('asc');
    }
  };

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    const list = q
      ? PROVIDERS.filter((p) => p.name.toLowerCase().includes(q))
      : PROVIDERS.slice();

    list.sort((a, b) => {
      const mul = sortDir === 'asc' ? 1 : -1;
      if (sortKey === 'status') {
        const delta = STATUS_RANK[a.status] - STATUS_RANK[b.status];
        if (delta !== 0) return delta * mul;
        return a.name.localeCompare(b.name) * mul;
      }
      return a.name.localeCompare(b.name) * mul;
    });
    return list;
  }, [query, sortKey, sortDir]);

  return (
    <div className={styles.root}>
      <div className={styles.toolbar}>
        <label className={styles.searchWrap}>
          <SearchIcon size={15} />
          <input
            type="text"
            placeholder="Filter providers"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className={styles.searchInput}
            aria-label="Filter providers by name"
          />
        </label>
        <span className={styles.count}>
          {filtered.length} / {PROVIDERS.length} providers
        </span>
      </div>

      {/* Desktop: sortable table */}
      <div className={styles.tableWrap}>
        <table className={styles.table}>
          <thead>
            <tr>
              <th scope="col" className={styles.sortable}>
                <button
                  type="button"
                  onClick={() => toggleSort('name')}
                  className={styles.sortButton}
                  aria-sort={
                    sortKey === 'name' ? (sortDir === 'asc' ? 'ascending' : 'descending') : 'none'
                  }>
                  Provider
                  <SortIcon active={sortKey === 'name'} dir={sortDir} />
                </button>
              </th>
              <th scope="col" className={styles.schemaHeader}>
                Schema config{' '}
                <span className={styles.headerAside}>
                  (<Link to="/docs/api/#aiservicebackendspec">AIServiceBackend</Link>)
                </span>
              </th>
              <th scope="col" className={styles.authHeader}>
                Auth{' '}
                <span className={styles.headerAside}>
                  (<Link to="/docs/api/#backendsecuritypolicyspec">BackendSecurityPolicy</Link>)
                </span>
              </th>
              <th scope="col" className={styles.sortable}>
                <button
                  type="button"
                  onClick={() => toggleSort('status')}
                  className={styles.sortButton}
                  aria-sort={
                    sortKey === 'status' ? (sortDir === 'asc' ? 'ascending' : 'descending') : 'none'
                  }>
                  Status
                  <SortIcon active={sortKey === 'status'} dir={sortDir} />
                </button>
              </th>
              <th scope="col">Note</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((p) => (
              <tr key={p.name}>
                <td className={styles.providerCell}>
                  {p.url ? (
                    <a href={p.url} target="_blank" rel="noreferrer noopener" className={styles.providerLink}>
                      {p.name}
                      <ExternalLinkIcon size={12} />
                    </a>
                  ) : (
                    <span>{p.name}</span>
                  )}
                </td>
                <td>
                  <div className={styles.schemaStack}>
                    {p.schemas.map((s) => (
                      <SchemaBlock key={`${s.label ?? 'default'}-${s.config.name}`} variant={s} />
                    ))}
                  </div>
                </td>
                <td>
                  <AuthList auth={p.auth} />
                </td>
                <td>
                  <StatusCell status={p.status} />
                </td>
                <td className={styles.noteCell}>{p.note ?? ''}</td>
              </tr>
            ))}
            {filtered.length === 0 && (
              <tr>
                <td colSpan={5} className={styles.empty}>
                  No providers match &ldquo;{query}&rdquo;.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Mobile: one card per provider */}
      <ul className={styles.cardList}>
        {filtered.map((p) => (
          <li key={p.name} className={styles.card}>
            <div className={styles.cardHeader}>
              {p.url ? (
                <a href={p.url} target="_blank" rel="noreferrer noopener" className={styles.providerLink}>
                  {p.name}
                  <ExternalLinkIcon size={12} />
                </a>
              ) : (
                <span className={styles.cardName}>{p.name}</span>
              )}
              <StatusCell status={p.status} />
            </div>

            <div className={styles.cardRow}>
              <span className={styles.cardLabel}>Schema</span>
              <div className={styles.schemaStack}>
                {p.schemas.map((s) => (
                  <SchemaBlock key={`${s.label ?? 'default'}-${s.config.name}`} variant={s} />
                ))}
              </div>
            </div>

            <div className={styles.cardRow}>
              <span className={styles.cardLabel}>Auth</span>
              <AuthList auth={p.auth} />
            </div>

            {p.note && (
              <div className={styles.cardRow}>
                <span className={styles.cardLabel}>Note</span>
                <p className={styles.cardNote}>{p.note}</p>
              </div>
            )}
          </li>
        ))}
        {filtered.length === 0 && (
          <li className={styles.empty}>No providers match &ldquo;{query}&rdquo;.</li>
        )}
      </ul>
    </div>
  );
}
