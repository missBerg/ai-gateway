import React from 'react';
import Link from '@docusaurus/Link';
import {
  Plug,
  Compass,
  Radio,
  Puzzle,
  Zap,
  Brain,
  Route,
  Target,
  Settings,
  TrendingUp,
  Copy,
  Repeat,
  TrafficCone,
  Scissors,
  Lock,
  Wrench,
  BarChart3,
  CheckCircle2,
  Package,
  Rocket,
  ShieldCheck,
  KeyRound,
  Gauge,
  FileCode,
  Terminal,
  Sparkles,
  BookOpen,
  Workflow,
  type LucideIcon,
} from 'lucide-react';
import styles from './styles.module.css';

// Map friendly icon names to Lucide components.
// Add new names here as needed.
const ICON_MAP: Record<string, LucideIcon> = {
  plug: Plug,
  compass: Compass,
  radio: Radio,
  puzzle: Puzzle,
  zap: Zap,
  brain: Brain,
  route: Route,
  target: Target,
  settings: Settings,
  'trending-up': TrendingUp,
  copy: Copy,
  repeat: Repeat,
  'traffic-cone': TrafficCone,
  scissors: Scissors,
  lock: Lock,
  wrench: Wrench,
  'bar-chart': BarChart3,
  check: CheckCircle2,
  package: Package,
  rocket: Rocket,
  'shield-check': ShieldCheck,
  key: KeyRound,
  gauge: Gauge,
  'file-code': FileCode,
  terminal: Terminal,
  sparkles: Sparkles,
  book: BookOpen,
  workflow: Workflow,
};

export type DocCard = {
  title: string;
  href: string;
  description?: string;
  /** Lucide icon name (e.g. "plug", "compass"). See ICON_MAP for supported names. */
  icon?: string;
};

interface DocCardGridProps {
  cards: DocCard[];
  columns?: 2 | 3 | 4;
}

const DocCardGrid: React.FC<DocCardGridProps> = ({ cards, columns = 3 }) => {
  const columnClass =
    columns === 2 ? styles.cols2 : columns === 4 ? styles.cols4 : styles.cols3;

  return (
    <div className={`${styles.grid} ${columnClass}`}>
      {cards.map((card) => {
        const IconComponent = card.icon ? ICON_MAP[card.icon] : null;
        return (
          <Link key={card.href} to={card.href} className={styles.card}>
            {IconComponent && (
              <span className={styles.icon} aria-hidden="true">
                <IconComponent size={22} strokeWidth={1.75} />
              </span>
            )}
            <span className={styles.title}>{card.title}</span>
            {card.description && (
              <span className={styles.description}>{card.description}</span>
            )}
            <span className={styles.arrow} aria-hidden="true">
              →
            </span>
          </Link>
        );
      })}
    </div>
  );
};

export default DocCardGrid;
