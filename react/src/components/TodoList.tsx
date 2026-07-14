import React, { useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useTodos, useFacets } from '../hooks/useTodos';
import { getCQRSMode, setCQRSMode } from '../api/todoApi';
import { useDebounce } from '../hooks/useDebounce';
import TodoItem from './TodoItem';
import TodoForm from './TodoForm';

const statuses = [
  { id: 1, name: 'Open', color: 'bg-blue-100 text-blue-800' },
  { id: 2, name: 'Progress', color: 'bg-yellow-100 text-yellow-800' },
  { id: 3, name: 'On Review', color: 'bg-purple-100 text-purple-800' },
  { id: 4, name: 'Done', color: 'bg-green-100 text-green-800' },
  { id: 5, name: 'Hold', color: 'bg-orange-100 text-orange-800' },
  { id: 6, name: 'Canceled', color: 'bg-gray-100 text-gray-800' },
];

const KanbanColumn: React.FC<{ status: number; name: string; color: string; search: string; facetCount: number }> = ({ status, name, color, search, facetCount }) => {
  const [page, setPage] = useState(1);
  const limit = 10; // Load 10 items per column per page
  
  // Reset page when search changes
  React.useEffect(() => {
    setPage(1);
  }, [search]);

  const { data, isLoading } = useTodos(search, status, page, limit);
  const todos = data?.data || [];
  const total = data?.total || 0;
  const totalPages = Math.ceil(total / limit);

  return (
    <div className="flex flex-col bg-gray-50/70 rounded-xl p-4 min-w-[320px] max-w-[350px] shrink-0 border border-gray-200 shadow-sm h-full max-h-[60vh]">
      <div className="flex items-center justify-between mb-4 pb-2 border-b border-gray-200">
        <h3 className={`px-3 py-1 rounded-full text-xs uppercase tracking-wider font-bold ${color}`}>
          {name}
        </h3>
        <span className="text-gray-600 text-xs font-bold bg-white px-3 py-1 rounded-full shadow-sm border border-gray-100">
          {facetCount || 0}
        </span>
      </div>
      
      <div className="flex-1 overflow-y-auto pr-1 custom-scrollbar">
        {isLoading ? (
          <div className="text-center py-8 text-sm text-gray-400 animate-pulse font-medium">Loading...</div>
        ) : todos.length === 0 ? (
          <div className="text-center py-8 text-sm text-gray-400 italic">No tasks here</div>
        ) : (
          <ul className="space-y-3">
            {todos.map((todo) => (
              <TodoItem key={todo.id} todo={todo} />
            ))}
          </ul>
        )}
      </div>

      {totalPages > 1 && (
        <div className="flex justify-between items-center mt-4 pt-3 border-t border-gray-200">
          <button
            onClick={() => setPage(p => Math.max(1, p - 1))}
            disabled={page === 1}
            className="text-xs text-gray-600 hover:text-indigo-600 disabled:opacity-30 disabled:hover:text-gray-600 font-semibold px-2 py-1 bg-white rounded shadow-sm border border-gray-200"
          >
            Prev
          </button>
          <span className="text-xs text-gray-500 font-medium bg-white px-2 py-1 rounded-md border border-gray-100 shadow-sm">
            {page} / {totalPages}
          </span>
          <button
            onClick={() => setPage(p => Math.min(totalPages, p + 1))}
            disabled={page === totalPages}
            className="text-xs text-gray-600 hover:text-indigo-600 disabled:opacity-30 disabled:hover:text-gray-600 font-semibold px-2 py-1 bg-white rounded shadow-sm border border-gray-200"
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
};

const TodoList: React.FC = () => {
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, 300);
  const [isCQRS, setIsCQRS] = useState(getCQRSMode());
  const queryClient = useQueryClient();

  const handleToggleCQRS = () => {
    const newMode = !isCQRS;
    setIsCQRS(newMode);
    setCQRSMode(newMode);
    // Invalidate caches to force a refetch using the new endpoint
    queryClient.invalidateQueries({ queryKey: ['todos'] });
    queryClient.invalidateQueries({ queryKey: ['facets'] });
  };

  const effectiveSearch = debouncedSearch.length >= 3 ? debouncedSearch : '';
  const { data: facets } = useFacets(effectiveSearch);

  return (
    <div className="max-w-[95vw] mx-auto p-6 bg-white/60 backdrop-blur-md rounded-2xl shadow-xl mt-6 min-h-[85vh] flex flex-col">
      <div className="flex flex-col items-center mb-8">
        <h1 className="text-4xl font-black text-gray-900 mb-2 text-center tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-indigo-600 to-purple-600">
          Task Master Kanban
        </h1>
        
        <div className="flex items-center gap-3 mt-2 bg-gray-50 px-5 py-2.5 rounded-full border border-gray-200 shadow-sm">
          <span className={`text-sm font-bold ${!isCQRS ? 'text-indigo-600' : 'text-gray-400'}`}>Standard (Raw)</span>
          <button
            onClick={handleToggleCQRS}
            className={`relative inline-flex h-7 w-14 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 shadow-inner ${
              isCQRS ? 'bg-purple-500' : 'bg-gray-300'
            }`}
            role="switch"
            aria-checked={isCQRS}
          >
            <span
              className={`inline-block h-5 w-5 transform rounded-full bg-white transition-transform shadow-md ${
                isCQRS ? 'translate-x-8' : 'translate-x-1'
              }`}
            />
          </button>
          <span className={`text-sm font-bold ${isCQRS ? 'text-purple-600' : 'text-gray-400'}`}>CQRS Mode</span>
        </div>
      </div>

      <div className="flex flex-col md:flex-row max-w-4xl mx-auto w-full mb-8 gap-4 items-start">
        <div className="flex-1 w-full">
          <div className="flex gap-2">
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search tasks (min 3 chars)..."
              className="w-full px-5 py-3 text-lg border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent shadow-sm transition-all"
            />
          </div>
          {search.length > 0 && search.length < 3 && (
            <div className="text-sm text-amber-600 mt-2 px-2 transition-all duration-300 ease-in-out font-medium">
              Please enter at least 3 characters to search.
            </div>
          )}
        </div>
        <div className="w-full md:w-1/2">
          <TodoForm />
        </div>
      </div>

      <div className="flex gap-6 overflow-x-auto pb-6 pt-2 flex-1 items-start px-4 -mx-4 scroll-smooth">
        {statuses.map(s => (
          <KanbanColumn
            key={s.id}
            status={s.id}
            name={s.name}
            color={s.color}
            search={effectiveSearch}
            facetCount={facets ? facets[s.id] || 0 : 0}
          />
        ))}
      </div>
    </div>
  );
};

export default TodoList;
