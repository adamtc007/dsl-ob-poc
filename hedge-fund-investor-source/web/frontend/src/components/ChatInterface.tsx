import React, { useState, useRef, useEffect } from "react";
import { ChatMessage } from "../types";

interface ChatInterfaceProps {
  messages: ChatMessage[];
  onSendMessage: (message: string) => void;
  isLoading: boolean;
}

const ChatInterface: React.FC<ChatInterfaceProps> = ({
  messages,
  onSendMessage,
  isLoading,
}) => {
  const [input, setInput] = useState("");
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (input.trim() && !isLoading) {
      onSendMessage(input.trim());
      setInput("");
    }
  };

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  // Focus input on component mount
  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  return (
    <div className="chat-container bg-gray-900 border border-gray-800">
      <div className="message-list bg-gray-900">
        {messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center text-gray-400">
            <p className="mb-3">No messages yet</p>
            <p className="text-sm max-w-md">
              Start a conversation with the Hedge Fund DSL Agent. Try asking:
              <br />
              <br />
              &quot;Create an opportunity for a Swiss corporate investor named
              Alpine Capital&quot;
              <br />
              <br />
              &quot;Begin KYC process for the investor&quot;
            </p>
          </div>
        ) : (
          messages.map((message, index) => (
            <div
              key={index}
              className={`message ${message.role === "user" ? "user-message" : "agent-message"}`}
            >
              <div className="mb-1 text-xs text-gray-400">
                {message.role === "user" ? "You" : "Agent"} •{" "}
                {new Date(message.timestamp).toLocaleTimeString()}
              </div>
              <div>{message.content}</div>
              {message.dsl && (
                <div className="mt-2 text-xs bg-gray-900 p-2 rounded border border-gray-700 font-mono overflow-x-auto">
                  <pre>{message.dsl}</pre>
                </div>
              )}
              {message.response?.confidence && (
                <div className="mt-1 text-xs text-gray-400">
                  Confidence: {(message.response.confidence * 100).toFixed(0)}%
                </div>
              )}
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>
      <form onSubmit={handleSubmit} className="input-area">
        <input
          type="text"
          ref={inputRef}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Type your message..."
          className="input-field"
          disabled={isLoading}
        />
        <button
          type="submit"
          className="send-button"
          disabled={isLoading || !input.trim()}
        >
          {isLoading ? (
            <span className="inline-block w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
          ) : (
            "Send"
          )}
        </button>
      </form>
    </div>
  );
};

export default ChatInterface;
