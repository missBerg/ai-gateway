import React from 'react';
import Heading from '@theme/Heading';
import Link from '@docusaurus/Link';
import { useLatestBlogs } from '@site/src/hooks/useLatestBlogs';
import type { BlogPostMeta } from '@site/src/plugins/latestBlogsPlugin';
import styles from './styles.module.css';

function formatDate(iso: string): string {
  if (!iso) return '';
  const [y, m, d] = iso.split('-').map(Number);
  if (!y || !m || !d) return iso;
  const date = new Date(Date.UTC(y, m - 1, d));
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    timeZone: 'UTC',
  });
}

function PostMeta({ date, readTime }: { date: string; readTime?: string }) {
  const pretty = formatDate(date);
  if (!pretty && !readTime) return null;
  return (
    <div className={styles.postMeta}>
      {pretty && <time dateTime={date}>{pretty}</time>}
      {pretty && readTime && <span className={styles.metaDot} aria-hidden="true">•</span>}
      {readTime && <span>{readTime}</span>}
    </div>
  );
}

function Tags({ tags }: { tags: string[] }) {
  if (tags.length === 0) return null;
  return (
    <div className={styles.tagsContainer}>
      {tags.slice(0, 3).map((tag, i) => (
        <span key={`${tag}-${i}`} className={styles.tag}>
          {tag}
        </span>
      ))}
    </div>
  );
}

function FeaturedCard({ post }: { post: BlogPostMeta }) {
  return (
    <Link to={post.permalink} className={styles.featuredCard} aria-label={post.title}>
      {post.image && (
        <div className={styles.featuredImageContainer}>
          <img
            src={post.image}
            alt=""
            className={styles.featuredImage}
            loading="lazy"
          />
        </div>
      )}
      <div className={styles.featuredBody}>
        <span className={styles.featuredEyebrow}>Featured</span>
        <Heading as="h3" className={styles.featuredTitle}>
          {post.title}
        </Heading>
        {post.description && (
          <p className={styles.featuredDescription}>{post.description}</p>
        )}
        <div className={styles.featuredFooter}>
          <PostMeta date={post.date} readTime={post.readTime} />
          <Tags tags={post.tags} />
        </div>
      </div>
    </Link>
  );
}

function BlogCard({ post }: { post: BlogPostMeta }) {
  return (
    <Link to={post.permalink} className={styles.blogCard} aria-label={post.title}>
      {post.image && (
        <div className={styles.imageContainer}>
          <img
            src={post.image}
            alt=""
            className={styles.blogImage}
            loading="lazy"
          />
        </div>
      )}
      <div className={styles.cardContent}>
        <PostMeta date={post.date} readTime={post.readTime} />
        <Heading as="h3" className={styles.blogTitle}>
          {post.title}
        </Heading>
        {post.description && (
          <p className={styles.blogDescription}>{post.description}</p>
        )}
        <Tags tags={post.tags} />
      </div>
    </Link>
  );
}

export default function LatestBlogs(): React.ReactElement | null {
  const { featuredPost, recentPosts } = useLatestBlogs();

  if (!featuredPost && recentPosts.length === 0) {
    return null;
  }

  return (
    <section className={styles.section}>
      <div className="container">
        <div className="sectionHeader">
          <span className="sectionEyebrow sectionEyebrow--purple">From the blog</span>
          <Heading as="h2" className="sectionTitle">
            Latest from the team
          </Heading>
          <p className="sectionSubtitle">
            News, deep-dives, and release notes from the Envoy AI Gateway maintainers.
          </p>
        </div>

        {featuredPost && <FeaturedCard post={featuredPost} />}

        {recentPosts.length > 0 && (
          <div className={styles.blogsGrid}>
            {recentPosts.map((post) => (
              <BlogCard key={post.slug} post={post} />
            ))}
          </div>
        )}

        <div className={styles.ctaSection}>
          <Link className="button button--primary button--lg" to="/blog">
            View all posts
          </Link>
        </div>
      </div>
    </section>
  );
}
