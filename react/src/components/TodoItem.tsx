import React from 'react';
import { Todo } from '../api/todoApi';
import { useUpdateTodo, useDeleteTodo } from '../hooks/useTodos';

interface TodoItemProps {
  todo: Todo;
}

const TodoItem: React.FC<TodoItemProps> = ({ todo }) => {
  const updateTodoMutation = useUpdateTodo();
  const deleteTodoMutation = useDeleteTodo();

  const handleStatusChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const requirePasskey = import.meta.env.VITE_REQUIRE_PASSKEY === 'true';
    const passkey = requirePasskey ? window.prompt('Enter admin passkey:') : '';
    if (passkey === null) return;

    updateTodoMutation.mutate(
      { todo: { ...todo, status: parseInt(e.target.value) }, passkey },
      {
        onError: () => {
          alert('Unauthorized: Invalid Passkey');
        },
      }
    );
  };

  const handleDelete = () => {
    const requirePasskey = import.meta.env.VITE_REQUIRE_PASSKEY === 'true';
    const passkey = requirePasskey ? window.prompt('Enter admin passkey:') : '';
    if (passkey === null) return;

    deleteTodoMutation.mutate(
      { id: todo.id, passkey },
      {
        onError: () => {
          alert('Unauthorized: Invalid Passkey');
        },
      }
    );
  };

  const handleDragStart = (e: React.DragEvent<HTMLLIElement>) => {
    e.dataTransfer.setData('application/json', JSON.stringify(todo));
    e.dataTransfer.effectAllowed = 'move';
  };

  return (
    <li
      draggable
      onDragStart={handleDragStart}
      className="flex flex-col p-4 mb-3 bg-white rounded-xl shadow-sm border border-gray-100 transition-all hover:shadow-md group cursor-grab active:cursor-grabbing"
    >
      <div className="flex items-start justify-between">
        <span
          className={`text-base font-medium ${todo.status === 6 || todo.status === 4 ? 'text-gray-400 line-through' : 'text-gray-800'
            }`}
        >
          {todo.title}
        </span>
        <button
          onClick={handleDelete}
          className="text-gray-300 hover:text-red-500 p-1 rounded-md transition-colors opacity-0 group-hover:opacity-100"
          aria-label="Delete todo"
        >
          <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clipRule="evenodd" />
          </svg>
        </button>
      </div>
      <div className="mt-3 flex justify-end">
        <select
          value={todo.status}
          onChange={handleStatusChange}
          disabled={updateTodoMutation.isPending}
          className="text-xs bg-gray-50 border border-gray-200 text-gray-700 rounded-md px-2 py-1 focus:outline-none focus:ring-1 focus:ring-indigo-500 cursor-pointer"
        >
          <option value={1}>Open</option>
          <option value={2}>Progress</option>
          <option value={3}>On Review</option>
          <option value={4}>Done</option>
          <option value={5}>Hold</option>
          <option value={6}>Canceled</option>
        </select>
      </div>
    </li>
  );
};

export default TodoItem;
