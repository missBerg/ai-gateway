import React from 'react';
import {Glyph} from '@site/src/components/icons';
import styles from './styles.module.css';

/*
 * Shared UI primitives for the Envoy AI Gateway site. These consume the global
 * design tokens (feedback family, focus-ring, shadows) defined in tokens.css.
 * Import individually, e.g. `import {Badge, Callout} from '@site/src/components/ui'`.
 */

type FeedbackVariant = 'info' | 'success' | 'warning' | 'danger';
type BadgeVariant = 'neutral' | 'brand' | 'accent' | FeedbackVariant;

export function Badge({
  children,
  variant = 'neutral',
  size = 'md',
  dot = false,
}: {
  children: React.ReactNode;
  variant?: BadgeVariant;
  size?: 'sm' | 'md';
  dot?: boolean;
}): React.ReactElement {
  return (
    <span className={styles.badge} data-variant={variant} data-size={size}>
      {dot && <span className={styles.badgeDot} />}
      {children}
    </span>
  );
}

export function Field({
  label,
  hint,
  error,
  required = false,
  children,
}: {
  label: string;
  hint?: string;
  error?: string;
  required?: boolean;
  children: React.ReactNode;
}): React.ReactElement {
  return (
    <label className={styles.field}>
      <span className={styles.fieldLabel}>
        {label}
        {required && <span className={styles.req}>*</span>}
      </span>
      {children}
      {error ? (
        <span className={styles.fieldError}>{error}</span>
      ) : (
        hint && <span className={styles.fieldHint}>{hint}</span>
      )}
    </label>
  );
}

/** Class helpers so consumers can style raw <input>/<select> consistently. */
export const inputClass = styles.input;
export const selectClass = `${styles.input} ${styles.select}`;

export function Toggle({
  label,
  defaultChecked = false,
  onChange,
}: {
  label: string;
  defaultChecked?: boolean;
  onChange?: (checked: boolean) => void;
}): React.ReactElement {
  return (
    <label className={styles.toggle}>
      <input
        type="checkbox"
        defaultChecked={defaultChecked}
        onChange={(e) => onChange?.(e.currentTarget.checked)}
      />
      <span className={styles.toggleTrack}>
        <span className={styles.toggleThumb} />
      </span>
      <span className={styles.toggleLabel}>{label}</span>
    </label>
  );
}

const CALLOUT_ICON: Record<FeedbackVariant, string> = {
  info: 'sparkles',
  success: 'check',
  warning: 'zap',
  danger: 'shield-check',
};

export function Callout({
  variant = 'info',
  title,
  children,
}: {
  variant?: FeedbackVariant;
  title?: string;
  children: React.ReactNode;
}): React.ReactElement {
  return (
    <div className={styles.callout} data-variant={variant}>
      <span className={styles.calloutIcon}>
        <Glyph name={CALLOUT_ICON[variant]} size={18} />
      </span>
      <div className={styles.calloutBody}>
        {title && <p className={styles.calloutTitle}>{title}</p>}
        <p className={styles.calloutText}>{children}</p>
      </div>
    </div>
  );
}

export function EmptyState({
  icon = 'compass',
  title,
  description,
  action,
}: {
  icon?: string;
  title: string;
  description: string;
  action?: React.ReactNode;
}): React.ReactElement {
  return (
    <div className={styles.empty}>
      <span className={styles.emptyIcon}>
        <Glyph name={icon} size={26} />
      </span>
      <p className={styles.emptyTitle}>{title}</p>
      <p className={styles.emptyText}>{description}</p>
      {action}
    </div>
  );
}

export function Stat({
  value,
  label,
  delta,
  direction,
}: {
  value: string;
  label: string;
  delta?: string;
  direction?: 'up' | 'down';
}): React.ReactElement {
  return (
    <div className={styles.stat}>
      <div className={styles.statValue}>{value}</div>
      <div className={styles.statLabel}>{label}</div>
      {delta && (
        <div className={styles.statDelta} data-dir={direction}>
          <Glyph name="bar-chart" size={13} />
          {delta}
        </div>
      )}
    </div>
  );
}

/** Grid wrapper for a row of <Stat> cards. */
export function StatGrid({children}: {children: React.ReactNode}): React.ReactElement {
  return <div className={styles.statGrid}>{children}</div>;
}

export function Kbd({children}: {children: React.ReactNode}): React.ReactElement {
  return <kbd className={styles.kbd}>{children}</kbd>;
}

export function Spinner(): React.ReactElement {
  return <span className={styles.spinner} aria-hidden="true" />;
}

export function Tooltip({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}): React.ReactElement {
  return (
    <span className={styles.tooltip} tabIndex={0}>
      {children}
      <span className={styles.tooltipBubble} role="tooltip">
        {label}
      </span>
    </span>
  );
}
