export default function Home() {
  return (
    <div style={{ padding: "2rem", fontFamily: "monospace" }}>
      <h1>Next.js OTel Example</h1>
      <p>Test endpoints:</p>
      <ul>
        <li>
          <a href="/nextjs/api/users">GET /nextjs/api/users</a>
        </li>
        <li>
          <a href="/nextjs/api/users/1">GET /nextjs/api/users/1</a>
        </li>
        <li>
          <a href="/nextjs/api/users/2">GET /nextjs/api/users/2</a>
        </li>
        <li>
          <a href="/nextjs/api/slow">GET /nextjs/api/slow</a> (manual span)
        </li>
        <li>
          <a href="/nextjs/api/external">GET /nextjs/api/external</a> (auto fetch span)
        </li>
        <li>
          <a href="/nextjs/api/test-error">GET /nextjs/api/test-error</a> (exception)
        </li>
      </ul>
    </div>
  );
}
