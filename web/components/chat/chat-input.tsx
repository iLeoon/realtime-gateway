'use client';

import { useState } from 'react';
import { useRef } from 'react';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

type ChatInputProps = {
  disabled?: boolean;
  onSend: (message: string) => void;
};

const SendIcon = () => (
  <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="1.8">
    <path d="M4 12h14" />
    <path d="M13 5l7 7-7 7" />
  </svg>
);

export const ChatInput = ({ disabled, onSend }: ChatInputProps) => {
  const [value, setValue] = useState('');
  const lastSubmitAtRef = useRef(0);

  const submit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const now = Date.now();
    if (now - lastSubmitAtRef.current < 250) {
      return;
    }
    lastSubmitAtRef.current = now;

    const next = value.trim();
    if (!next || disabled) {
      return;
    }

    onSend(next);
    setValue('');
  };

  return (
    <form
      onSubmit={submit}
      className="sticky bottom-0 z-20 border-t border-border/70 bg-background/80 px-4 py-3 backdrop-blur-md sm:px-5"
    >
      <div className="flex items-center gap-2 rounded-full border border-border/70 bg-muted/40 p-1.5 shadow-sm">
        <Input
          value={value}
          onChange={(event) => setValue(event.target.value)}
          placeholder="Type a message"
          disabled={disabled}
          className="h-10 border-none bg-transparent shadow-none focus-visible:ring-0 focus-visible:ring-offset-0"
        />

        <Button
          type="submit"
          size="sm"
          className="h-9 w-9 rounded-full p-0 shadow-sm transition-all duration-200 hover:scale-105"
          disabled={disabled || value.trim().length === 0}
        >
          <SendIcon />
        </Button>
      </div>
    </form>
  );
};
