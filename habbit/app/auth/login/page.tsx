'use client';
import { useState } from 'react';

export default function Login() {
  const [email,setEmail] = useState('');
  const [password,setPassword] = useState('');
  const [error,setError] = useState('');

  async function submit(e:any){
    e.preventDefault();
    setError('');
    const res = await fetch('/api/auth/login', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ email, password }) });
    if (!res.ok) { setError('Invalid credentials or email not verified'); return; }
    window.location.href = '/dashboard';
  }

  return (
    <div className="card">
      <h2>Login</h2>
      <form onSubmit={submit} className="grid">
        <input placeholder="Email" value={email} onChange={e=>setEmail(e.target.value)} />
        <input placeholder="Password" type="password" value={password} onChange={e=>setPassword(e.target.value)} />
        {error && <div>{error}</div>}
        <button type="submit">Login</button>
      </form>
    </div>
  );
}
