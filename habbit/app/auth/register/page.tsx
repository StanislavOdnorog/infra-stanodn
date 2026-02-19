'use client';
import { useState } from 'react';

export default function Register() {
  const [email,setEmail] = useState('');
  const [password,setPassword] = useState('');
  const [message,setMessage] = useState('');

  async function submit(e:any){
    e.preventDefault();
    setMessage('');
    const res = await fetch('/api/auth/register', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ email, password }) });
    if (!res.ok) { setMessage('Registration failed'); return; }
    setMessage('Check your email to verify.');
  }

  return (
    <div className="card">
      <h2>Register</h2>
      <form onSubmit={submit} className="grid">
        <input placeholder="Email" value={email} onChange={e=>setEmail(e.target.value)} />
        <input placeholder="Password" type="password" value={password} onChange={e=>setPassword(e.target.value)} />
        {message && <div>{message}</div>}
        <button type="submit">Create account</button>
      </form>
    </div>
  );
}
