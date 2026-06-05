import React, {useEffect, useRef, useState} from 'react';
import styles from './styles.module.css';

type Props = {
  children: React.ReactNode;
  /** Stagger index for sequenced group reveals (0, 1, 2…). */
  index?: number;
  /** Travel distance in px (default 24). */
  distance?: number;
  /** Render as a different element (default div). */
  as?: keyof React.JSX.IntrinsicElements;
  className?: string;
};

/**
 * Floats children up + fades them in when scrolled into view.
 *
 * Standard, universally-supported pattern: IntersectionObserver flips a
 * data-reveal attribute and CSS handles the transition.
 *  - SSR-safe: renders visible (data-reveal unset → opacity 1), so no-JS and
 *    prerendered output never get stuck hidden.
 *  - On mount the element arms (hidden) only if motion is allowed; reduced-motion
 *    users skip straight to visible with no transform.
 */
export default function Reveal({
  children,
  index = 0,
  distance = 24,
  as = 'div',
  className,
}: Props): React.ReactElement {
  const ref = useRef<HTMLElement | null>(null);
  const [state, setState] = useState<'idle' | 'armed' | 'in'>('idle');

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    const reduce = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches;
    if (reduce || typeof IntersectionObserver === 'undefined') {
      setState('in'); // visible, no motion
      return;
    }

    setState('armed'); // hide, ready to float in
    const io = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            setState('in');
            io.unobserve(entry.target);
          }
        }
      },
      {threshold: 0.15, rootMargin: '0px 0px -8% 0px'},
    );
    io.observe(el);
    return () => io.disconnect();
  }, []);

  const Tag = as as React.ElementType;
  return (
    <Tag
      ref={ref}
      className={`${styles.reveal} ${className ?? ''}`}
      data-reveal={state === 'idle' ? undefined : state}
      style={{
        '--reveal-i': index,
        '--reveal-distance': `${distance}px`,
      } as React.CSSProperties}>
      {children}
    </Tag>
  );
}
