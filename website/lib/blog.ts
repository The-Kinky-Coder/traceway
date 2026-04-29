import fs from "node:fs";
import path from "node:path";
import matter from "gray-matter";
import { Marked } from "marked";
import { markedHighlight } from "marked-highlight";
import hljs from "highlight.js";

export type PostMeta = {
  slug: string;
  title: string;
  date: string;
  author?: string;
  excerpt?: string;
  tags?: string[];
};

export type Post = PostMeta & {
  content: string;
};

const BLOG_DIR = path.join(process.cwd(), "content", "blog");

const marked = new Marked(
  markedHighlight({
    langPrefix: "hljs language-",
    highlight(code, lang) {
      const language = lang && hljs.getLanguage(lang) ? lang : "plaintext";
      return hljs.highlight(code, { language }).value;
    },
  }),
);

marked.setOptions({ gfm: true, breaks: false });

function readPostFile(slug: string): { data: matter.GrayMatterFile<string>["data"]; content: string } {
  const file = path.join(BLOG_DIR, `${slug}.md`);
  const raw = fs.readFileSync(file, "utf8");
  const parsed = matter(raw);
  return { data: parsed.data, content: parsed.content };
}

function toMeta(slug: string, data: Record<string, unknown>): PostMeta {
  const date = data.date instanceof Date ? data.date.toISOString() : String(data.date ?? "");
  return {
    slug,
    title: String(data.title ?? slug),
    date,
    author: data.author ? String(data.author) : undefined,
    excerpt: data.excerpt ? String(data.excerpt) : undefined,
    tags: Array.isArray(data.tags) ? data.tags.map(String) : undefined,
  };
}

export function getAllPostSlugs(): string[] {
  if (!fs.existsSync(BLOG_DIR)) return [];
  return fs
    .readdirSync(BLOG_DIR)
    .filter((f) => f.endsWith(".md"))
    .map((f) => f.replace(/\.md$/, ""));
}

export function getAllPosts(): PostMeta[] {
  return getAllPostSlugs()
    .map((slug) => {
      const { data } = readPostFile(slug);
      return toMeta(slug, data);
    })
    .sort((a, b) => (a.date < b.date ? 1 : -1));
}

export function getPostBySlug(slug: string): Post | null {
  if (!fs.existsSync(path.join(BLOG_DIR, `${slug}.md`))) return null;
  const { data, content } = readPostFile(slug);
  return { ...toMeta(slug, data), content };
}

export async function markdownToHtml(md: string): Promise<string> {
  return await marked.parse(md);
}

export function formatPostDate(iso: string): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}
