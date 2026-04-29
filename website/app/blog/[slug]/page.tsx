import Link from "next/link";
import { notFound } from "next/navigation";
import type { Metadata } from "next";
import { ArrowLeft } from "lucide-react";
import { Eyebrow } from "@/components/eyebrow";
import { getAllPostSlugs, getPostBySlug, markdownToHtml, formatPostDate } from "@/lib/blog";

type Params = { slug: string };

export function generateStaticParams(): Params[] {
  return getAllPostSlugs().map((slug) => ({ slug }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<Params>;
}): Promise<Metadata> {
  const { slug } = await params;
  const post = getPostBySlug(slug);
  if (!post) return { title: "Not found — Traceway Blog" };
  return {
    title: `${post.title} — Traceway Blog`,
    description: post.excerpt,
    alternates: { canonical: `/blog/${post.slug}` },
  };
}

export default async function BlogPostPage({
  params,
}: {
  params: Promise<Params>;
}) {
  const { slug } = await params;
  const post = getPostBySlug(slug);
  if (!post) notFound();

  const html = await markdownToHtml(post.content);

  return (
    <main className="relative">
      <article className="py-20">
        <div className="wrap max-w-3xl mx-auto">
          <Link
            href="/blog"
            className="inline-flex items-center gap-1.5 text-[12px] tracking-[.06em] uppercase transition-colors hover:text-[color:var(--a2)]"
            style={{ fontFamily: "var(--font-mono)", color: "var(--fg-3)" }}
          >
            <ArrowLeft className="h-3 w-3" />
            Back to blog
          </Link>

          <header className="mt-8">
            <Eyebrow>Post</Eyebrow>
            <h1 className="mt-4">{post.title}</h1>
            <div
              className="mt-5 text-[12px] tracking-[.06em] uppercase"
              style={{ fontFamily: "var(--font-mono)", color: "var(--fg-3)" }}
            >
              <time dateTime={post.date}>{formatPostDate(post.date)}</time>
              {post.author ? <span> · {post.author}</span> : null}
              {post.tags && post.tags.length > 0 ? (
                <span> · {post.tags.join(", ")}</span>
              ) : null}
            </div>
          </header>

          <div
            className="prose-tw mt-12"
            dangerouslySetInnerHTML={{ __html: html }}
          />
        </div>
      </article>
    </main>
  );
}
