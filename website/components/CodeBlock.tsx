"use client";

interface CodeBlockProps {
  code: string;
  filename?: string;
}

export default function CodeBlock({ code, filename }: CodeBlockProps) {
  return (
    <div className="rounded-lg border border-border overflow-hidden my-6">
      {filename && (
        <div className="bg-bg-card px-4 py-2 border-b border-border">
          <span className="font-mono text-xs text-text-muted">{filename}</span>
        </div>
      )}
      <pre className="bg-[#0d0d0d] p-4 overflow-x-auto">
        <code className="font-mono text-sm text-text leading-relaxed whitespace-pre">
          {code}
        </code>
      </pre>
    </div>
  );
}
