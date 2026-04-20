import type { LoadContext, Plugin } from '@docusaurus/types';
import * as fs from 'fs';
import * as path from 'path';
import matter from 'gray-matter';
import { glob } from 'glob';

export type BlogPostMeta = {
  title: string;
  description: string;
  image: string;
  tags: string[];
  slug: string;
  date: string;
  permalink: string;
  readTime?: string;
  featured?: boolean;
};

export type LatestBlogsPluginData = {
  featuredPost: BlogPostMeta | null;
  recentPosts: BlogPostMeta[];
  /** @deprecated kept for backwards compat; prefer featuredPost + recentPosts */
  latestPosts: BlogPostMeta[];
};

const WORDS_PER_MINUTE = 220;

function estimateReadTime(body: string): string {
  const words = body.trim().split(/\s+/).filter(Boolean).length;
  if (words === 0) return '';
  const minutes = Math.max(1, Math.round(words / WORDS_PER_MINUTE));
  return `${minutes} min read`;
}

export default function latestBlogsPlugin(context: LoadContext): Plugin {
  return {
    name: 'docusaurus-plugin-latest-blogs',

    async contentLoaded({ actions }) {
      const { setGlobalData } = actions;
      const blogDir = path.join(context.siteDir, 'blog');

      const files = await glob('**/*.{md,mdx}', { cwd: blogDir });

      const posts: BlogPostMeta[] = files
        .map((file) => {
          try {
            const filePath = path.join(blogDir, file);
            const content = fs.readFileSync(filePath, 'utf-8');
            const { data, content: body } = matter(content);

            if (!data.title || !data.slug) {
              return null;
            }

            const dateMatch = file.match(/(\d{4}-\d{2}-\d{2})/);
            const date = dateMatch?.[1] || '';

            const readTime: string =
              typeof data.readTime === 'string' && data.readTime.length > 0
                ? data.readTime
                : estimateReadTime(body);

            return {
              title: data.title,
              description: data.description || '',
              image: data.image || '',
              tags: data.tags || [],
              slug: data.slug,
              date,
              permalink: `/blog/${data.slug}`,
              readTime,
              featured: Boolean(data.featured),
            };
          } catch {
            return null;
          }
        })
        .filter((post): post is BlogPostMeta => post !== null)
        .sort((a, b) => b.date.localeCompare(a.date));

      const featuredPost = posts.find((p) => p.featured) ?? null;
      const recentPosts = posts
        .filter((p) => p.permalink !== featuredPost?.permalink)
        .slice(0, 3);

      const latestPosts = posts.slice(0, 3);

      setGlobalData({
        featuredPost,
        recentPosts,
        latestPosts,
      } as LatestBlogsPluginData);
    },
  };
}
