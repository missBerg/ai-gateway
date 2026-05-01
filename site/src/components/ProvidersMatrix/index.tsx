import React, {useCallback, useMemo, useState} from 'react';
import Link from '@docusaurus/Link';
import {
  CheckCircle2,
  AlertTriangle,
  Clock,
  ChevronUp,
  ChevronDown,
  ChevronsUpDown,
  Copy,
  Check,
  Search,
  ExternalLink,
} from 'lucide-react';
import {
  AUTH_META,
  PROVIDERS,
  type Provider,
  type ProviderStatus,
  type SchemaVariant,
} from './providers';
import styles from './styles.module.css';

type SortKey = 'name' | 'status';
type SortDir = 'asc' | 'desc';

// Status ordering — supported first, then partial, then anything planned.
const STATUS_RANK: Record<ProviderStatus, number> = {
  supported: 0,
  partial: 1,
  planned: 2,
};

const STATUS_META: Record<
  ProviderStatus,
  {Icon: typeof CheckCircle2; label: string; cls: string}
> = {
  supported: {Icon: CheckCircle2, label: 'Supported', cls: styles.statusOk},
  partial: {
    Icon: AlertTriangle,
    label: 'Partial',
    cls: styles.statusWarn,
  },
  planned: {Icon: Clock, label: 'Planned', cls: styles.statusPlanned},
};

function StatusCell({status}: {status: ProviderStatus}) {
  const {Icon, label, cls} = STATUS_META[status];
  return (
    <span className={`${styles.status} ${cls}`} title={label}>
      <Icon size={16} strokeWidth={2} aria-hidden="true" />
      <span className={styles.statusLabel}>{label}</span>
    </span>
  );
}

function formatConfig(variant: SchemaVariant): string {
  // Pretty-print for display and copy. Two-space indent matches the docs'
  // YAML/JSON conventions.
  return JSON.stringify(variant.config, null, 2);
}

function SchemaBlock({variant}: {variant: SchemaVariant}) {
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
      {variant.label && (
        <span className={styles.schemaLabel}>{variant.label}</span>
      )}
      <div className={styles.codeWrap}>
        <button
          type="button"
          className={styles.copyBtn}
          onClick={handleCopy}
          aria-label={copied ? 'Copied' : 'Copy schema config'}
        >
          {copied ? (
            <Check size={13} strokeWidth={2.4} aria-hidden="true" />
          ) : (
            <Copy size={13} strokeWidth={2} aria-hidden="true" />
          )}
        </button>
        <pre className={styles.codeBlock}>
          <code>{text}</code>
        </pre>
      </div>
    </div>
  );
}

function AuthList({auth}: {auth: Provider['auth']}) {
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

function SortIcon({active, dir}: {active: boolean; dir: SortDir}) {
  if (!active) {
    return (
      <ChevronsUpDown
        size={13}
        strokeWidth={2}
        className={styles.sortIcon}
        aria-hidden="true"
      />
    );
  }
  return dir === 'asc' ? (
    <ChevronUp
      size={13}
      strokeWidth={2.2}
      className={`${styles.sortIcon} ${styles.sortIconActive}`}
      aria-hidden="true"
    />
  ) : (
    <ChevronDown
      size={13}
      strokeWidth={2.2}
      className={`${styles.sortIcon} ${styles.sortIconActive}`}
      aria-hidden="true"
    />
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
        // Secondary sort by name keeps order stable within a status.
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
          <Search size={15} strokeWidth={2} aria-hidden="true" />
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
                    sortKey === 'name'
                      ? sortDir === 'asc'
                        ? 'ascending'
                        : 'descending'
                      : 'none'
                  }
                >
                  Provider
                  <SortIcon active={sortKey === 'name'} dir={sortDir} />
                </button>
              </th>
              <th scope="col" className={styles.schemaHeader}>
                Schema config{' '}
                <span className={styles.headerAside}>
                  (
                  <Link to="/docs/api/#aiservicebackendspec">
                    AIServiceBackend
                  </Link>
                  )
                </span>
              </th>
              <th scope="col" className={styles.authHeader}>
                Auth{' '}
                <span className={styles.headerAside}>
                  (
                  <Link to="/docs/api/#backendsecuritypolicyspec">
                    BackendSecurityPolicy
                  </Link>
                  )
                </span>
              </th>
              <th scope="col" className={styles.sortable}>
                <button
                  type="button"
                  onClick={() => toggleSort('status')}
                  className={styles.sortButton}
                  aria-sort={
                    sortKey === 'status'
                      ? sortDir === 'asc'
                        ? 'ascending'
                        : 'descending'
                      : 'none'
                  }
                >
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
                    <a
                      href={p.url}
                      target="_blank"
                      rel="noreferrer noopener"
                      className={styles.providerLink}
                    >
                      {p.name}
                      <ExternalLink
                        size={12}
                        strokeWidth={2}
                        aria-hidden="true"
                      />
                    </a>
                  ) : (
                    <span>{p.name}</span>
                  )}
                </td>
                <td>
                  <div className={styles.schemaStack}>
                    {p.schemas.map((s) => (
                      <SchemaBlock
                        key={`${s.label ?? 'default'}-${s.config.name}`}
                        variant={s}
                      />
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

      {/* Mobile: card per provider */}
      <ul className={styles.cardList}>
        {filtered.map((p) => (
          <li key={p.name} className={styles.card}>
            <div className={styles.cardHeader}>
              {p.url ? (
                <a
                  href={p.url}
                  target="_blank"
                  rel="noreferrer noopener"
                  className={styles.providerLink}
                >
                  {p.name}
                  <ExternalLink
                    size={12}
                    strokeWidth={2}
                    aria-hidden="true"
                  />
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
                  <SchemaBlock
                    key={`${s.label ?? 'default'}-${s.config.name}`}
                    variant={s}
                  />
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
          <li className={styles.empty}>
            No providers match &ldquo;{query}&rdquo;.
          </li>
        )}
      </ul>
    </div>
  );
}
