import React, { useState, useRef } from 'react';
import { useCreateTodo } from '../hooks/useTodos';

const TodoForm: React.FC = () => {
  const [title, setTitle] = useState('');
  const createTodoMutation = useCreateTodo();
  const inputRef = useRef<HTMLInputElement>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;

    const requirePasskey = import.meta.env.VITE_REQUIRE_PASSKEY === 'true';
    const passkey = requirePasskey ? window.prompt('Enter admin passkey:') : '';
    if (passkey === null) return; // User cancelled

    createTodoMutation.mutate({ title, passkey }, {
      onSuccess: () => {
        setTitle(''); // Clear input on success
        setTimeout(() => {
          inputRef.current?.focus();
        }, 0);
      },
      onError: () => {
        alert('Unauthorized: Invalid Passkey');
      },
    });
  };

  return (
    <form onSubmit={handleSubmit} className="flex gap-2 h-full">
      <input
        ref={inputRef}
        type="text"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="What needs to be done?"
        className="flex-1 px-5 py-3 text-lg border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 shadow-sm transition-shadow"
        disabled={createTodoMutation.isPending}
      />
      <button
        type="submit"
        disabled={createTodoMutation.isPending || !title.trim()}
        className="px-6 py-3 bg-indigo-600 text-white rounded-xl font-medium hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors shadow-sm"
      >
        {createTodoMutation.isPending ? 'Adding...' : 'Add Task'}
      </button>
    </form>
  );
};

export default TodoForm;
