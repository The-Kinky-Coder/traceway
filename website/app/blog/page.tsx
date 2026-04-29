import Link from "next/link";
import type { Metadata } from "next";
import { ArrowRight } from "lucide-react";
import { Eyebrow } from "@/components/eyebrow";
import { getAllPosts, formatPostDate } from "@/lib/blog";

export const metadata: Metadata = {
  title: "Engineering Blog — Traceway",
  description:
    "Release notes, engineering deep-dives, and design decisions behind Traceway — observability for modern backends.",
  alternates: { canonical: "/blog" },
};

export default function BlogIndex() {
  const posts = getAllPosts();

  return (
    <main className="relative">
      <section className="py-24 gridbg">
        <div className="wrap max-w-3xl mx-auto">
          <div className="text-center">
            <Eyebrow>Engineering Blog</Eyebrow>
            <h1 className="mt-6">Notes from building Traceway.</h1>
            <p className="mt-4 muted text-[15px] max-w-xl mx-auto">
              Release notes, architecture deep-dives, and the engineering decisions behind the
              platform.
            </p>
          </div>
        </div>
      </section>

      <section className="pb-24">
        <div className="wrap max-w-3xl mx-auto">
          {posts.length === 0 ? (
            <p className="muted text-center py-12">No posts yet — check back soon.</p>
          ) : (
            <ul className="flex flex-col gap-3">
              {posts.map((post) => (
                <li key={post.slug}>
                  <Link
                    href={`/blog/${post.slug}`}
                    className="group block rounded-[10px] p-6 transition-colors"
                    style={{
                      background: "var(--ink-2)",
                      border: "1px solid var(--hair)",
                    }}
                  >
                    <div
                      className="text-[11px] tracking-[.08em] uppercase mb-3"
                      style={{ fontFamily: "var(--font-mono)", color: "var(--fg-3)" }}
                    >
                      <time dateTime={post.date}>{formatPostDate(post.date)}</time>
                      {post.author ? <span> · {post.author}</span> : null}
                    </div>
                    <h2
                      className="text-[22px] font-semibold leading-tight transition-colors group-hover:text-[color:var(--a2)]"
                      style={{ fontFamily: "var(--font-display)", color: "var(--fg-0)", letterSpacing: "-0.01em" }}
                    >
                      {post.title}
                    </h2>
                    {post.excerpt ? (
                      <p
                        className="mt-2 text-[14px] leading-relaxed"
                        style={{ color: "var(--fg-2)" }}
                      >
                        {post.excerpt}
                      </p>
                    ) : null}
                    <span
                      className="mt-4 inline-flex items-center gap-1.5 text-[12px] tracking-[.06em] uppercase transition-colors group-hover:text-[color:var(--a2)]"
                      style={{ fontFamily: "var(--font-mono)", color: "var(--fg-2)" }}
                    >
                      Read post
                      <ArrowRight className="h-3 w-3 transition-transform group-hover:translate-x-0.5" />
                    </span>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </div>
      </section>
    </main>
  );
}
