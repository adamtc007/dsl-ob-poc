import React from "react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { vsDark } from "react-syntax-highlighter/dist/esm/styles/prism";

interface DSLViewerProps {
  code: string;
}

const customStyle = {
  ...vsDark,
  'pre[class*="language-"]': {
    ...vsDark['pre[class*="language-"]'],
    backgroundColor: "#1a1a2e",
    border: "1px solid #2a2a3e",
    borderRadius: "0.5rem",
    padding: "1rem",
    fontSize: "0.9rem",
    lineHeight: "1.5",
    overflow: "auto",
    maxHeight: "calc(100vh - 20rem)",
  },
  'code[class*="language-"]': {
    ...vsDark['code[class*="language-"]'],
    fontFamily: "'Fira Code', 'Courier New', monospace",
  },
};

// Custom language definition for the DSL
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const dslLanguage = {
  comment: {
    pattern: /#.*/,
    greedy: true,
  },
  string: {
    pattern: /"(?:\\.|[^"\\])*"/,
    greedy: true,
  },
  keyword:
    /\b(investor|kyc|tax|screen|subscribe|redeem|approve|reject|cash|nav)\b/,
  function:
    /\b(start-opportunity|begin-kyc|screen|collect-doc|approve|set-refresh-schedule|capture|set-instruction|request|confirm|issue|settle|close)\b/,
  "class-name": {
    pattern: /@attr\{[^}]*}/,
    greedy: true,
  },
  punctuation: /[()[\]{};=]/,
  operator: /\./,
  number: /\b\d+(?:\.\d+)?\b/,
  boolean: /\b(?:true|false)\b/,
};

const DSLViewer: React.FC<DSLViewerProps> = ({ code }) => {
  return (
    <div className="dsl-viewer-container">
      <SyntaxHighlighter
        language="lisp"
        style={customStyle}
        customStyle={{
          height: "calc(100vh - 12rem)",
          margin: 0,
        }}
        showLineNumbers={true}
        wrapLines={true}
      >
        {code}
      </SyntaxHighlighter>
    </div>
  );
};

export default DSLViewer;
