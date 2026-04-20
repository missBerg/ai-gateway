import { usePluginData } from '@docusaurus/useGlobalData';
import type {
  BlogPostMeta,
  LatestBlogsPluginData,
} from '../plugins/latestBlogsPlugin';

export function useLatestBlogs(): {
  featuredPost: BlogPostMeta | null;
  recentPosts: BlogPostMeta[];
} {
  const data = usePluginData('docusaurus-plugin-latest-blogs') as
    | LatestBlogsPluginData
    | undefined;
  return {
    featuredPost: data?.featuredPost ?? null,
    recentPosts: data?.recentPosts ?? [],
  };
}
