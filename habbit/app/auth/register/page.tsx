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
    const data = await res.json().catch(()=>({}));
    if (!res.ok) { setMessage(data.error || 'Registration failed'); return; }
    setMessage('Check your email to verify.');
  }

  return (
    <div className="card">
      <h2>Register</h2>
      <form onSubmit={submit} className="grid">
        <input placeholder="Email" value={email} onChange={e=>setEmail(e.target.value)} required />
        <input placeholder="Password (min 8)" type="password" value={password} onChange={e=>setPassword(e.target.value)} required />
        {message && <div className="alert">{message}</div>}
        <button type="submit" className="btn btn-primary">Create account</button>
      </form>
    </div>
  );
}
