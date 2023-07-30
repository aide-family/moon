import SyntaxHighlighter from "react-syntax-highlighter";
import React from "react";
import { docco } from "react-syntax-highlighter/dist/esm/styles/hljs";
import { Button, Message } from "@arco-design/web-react";
import { IconCopy } from "@arco-design/web-react/icon";

export interface CodeViewProps {
  codeString: string;
  language: string;
}

const CodeView: React.FC<CodeViewProps> = (props) => {
  const { codeString, language } = props;
  const copyCodeString = () => {
    window?.navigator?.clipboard?.writeText(codeString);
    Message.success("复制成功");
  };

  return (
    <div
      style={{
        position: "relative",
      }}
    >
      <Button
        icon={<IconCopy />}
        type="text"
        onClick={copyCodeString}
        style={{
          position: "absolute",
          top: 0,
          right: 0,
        }}
      />
      <SyntaxHighlighter language={language} style={docco}>
        {codeString}
      </SyntaxHighlighter>
    </div>
  );
};

export default CodeView;
