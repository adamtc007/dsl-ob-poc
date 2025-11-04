import { useState, useEffect, useRef } from 'react';
import ChatInterface from './components/ChatInterface';
import DSLViewer from './components/DSLViewer';
import { ChatMessage, DSLGenerationResponse } from './types';

const App = () => {
  const [sessionId, setSessionId] = useState<string>('');
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isConnected, setIsConnected] = useState<boolean>(false);
  const [currentState, setCurrentState] = useState<string>('');
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    // Initialize WebSocket connection
    const ws = new WebSocket(`${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`);

    ws.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      setIsConnected(false);
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);

      if (data.type === 'welcome') {
        setSessionId(data.payload.session_id);
        setMessages(prev => [...prev, {
          role: 'agent',
          content: data.payload.message,
          timestamp: new Date(),
        }]);
      } else if (data.type === 'chat_response') {
        setIsLoading(false);
        const response = data.payload;

        if (response.state) {
          setCurrentState(response.state);
        }

        setMessages(prev => [...prev, {
          role: 'agent',
          content: response.message,
          dsl: response.dsl,
          response: response.response as DSLGenerationResponse,
          timestamp: new Date(),
        }]);
      } else if (data.type === 'error') {
        setIsLoading(false);
        setMessages(prev => [...prev, {
          role: 'agent',
          content: `Error: ${data.payload.error}`,
          timestamp: new Date(),
        }]);
      }
    };

    wsRef.current = ws;

    // Clean up on unmount
    return () => {
      ws.close();
    };
  }, []);

  const sendMessage = (message: string) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.error('WebSocket not connected');
      return;
    }

    setIsLoading(true);
    setMessages(prev => [...prev, {
      role: 'user',
      content: message,
      timestamp: new Date(),
    }]);

    wsRef.current.send(JSON.stringify({
      type: 'chat',
      payload: {
        message,
        context: { session_id: sessionId }
      }
    }));
  };

  const getStateClass = () => {
    if (!currentState) return '';
    return `state-${currentState.toLowerCase()}`;
  };

  // Find the latest DSL message
  const latestDSL = messages
    .filter(msg => msg.dsl)
    .sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime())[0]?.dsl;

  return (
    <div className="flex flex-col h-full">
      <header className="mb-6">
        <h1 className="text-3xl font-bold">Hedge Fund DSL Agent</h1>
        <div className={`state-indicator ${getStateClass()}`}>
          <div className="state-circle"></div>
          <span className="state-name">{currentState || 'Not Started'}</span>
          {isConnected ?
            <span className="text-green-400 text-xs ml-auto">Connected</span> :
            <span className="text-red-400 text-xs ml-auto">Disconnected</span>
          }
        </div>
      </header>

      <div className="flex flex-col lg:flex-row gap-4 h-full">
        <div className="flex-1">
          <ChatInterface
            messages={messages}
            onSendMessage={sendMessage}
            isLoading={isLoading}
          />
        </div>

        <div className="w-full lg:w-2/5">
          <h2 className="text-xl font-semibold mb-3">Generated DSL</h2>
          <DSLViewer code={latestDSL || '# No DSL generated yet\n# Start a conversation to generate DSL'} />
        </div>
      </div>

      <footer className="mt-8 text-center text-sm text-gray-400">
        <p>Hedge Fund DSL Agent &copy; {new Date().getFullYear()}</p>
      </footer>
    </div>
  );
};

export default App;
