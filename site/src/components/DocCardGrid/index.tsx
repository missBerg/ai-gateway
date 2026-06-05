import React from 'react';
import Link from '@docusaurus/Link';
import {Glyph, hasGlyph} from '@site/src/components/icons';
import styles from './styles.module.css';

export type DocCard = {
  title: string;
  href: string;
  description?: string;
  /** Inline icon name (e.g. "plug", "rocket"). See src/components/icons.tsx. */
  icon?: string;
};

interface DocCardGridProps {
  cards: DocCard[];
  columns?: 2 | 3 | 4;
}

export default function DocCardGrid({cards, columns = 3}: DocCardGridProps): React.ReactElement {
  const columnClass =
    columns === 2 ? styles.cols2 : columns === 4 ? styles.cols4 : styles.cols3;

  return (
    <div className={`${styles.grid} ${columnClass}`}>
      {cards.map((card) => (
        <Link key={card.href} to={card.href} className={styles.card}>
          {hasGlyph(card.icon) && (
            <span className={styles.icon} aria-hidden="true">
              <Glyph name={card.icon!} size={22} strokeWidth={1.75} />
            </span>
          )}
          <span className={styles.title}>{card.title}</span>
          {card.description && <span className={styles.description}>{card.description}</span>}
          <span className={styles.arrow} aria-hidden="true">
            →
          </span>
        </Link>
      ))}
    </div>
  );
}
