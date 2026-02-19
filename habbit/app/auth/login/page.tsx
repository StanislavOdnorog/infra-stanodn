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
    if (res.status === 401) { setError('Invalid credentials'); return; }
    if (res.status === 403) { setError('Email not verified'); return; }
    if (!res.ok) { setError('Login failed'); return; }
    window.location.href = '/dashboard';
  }

  return (
    <div className="card">
      <h2>Login</h2>
      <form onSubmit={submit} className="grid">
        <input placeholder="Email" value={email} onChange={e=>setEmail(e.target.value)} required />
        <input placeholder="Password" type="password" value={password} onChange={e=>setPassword(e.target.value)} required />
        {error && <div className="alert">{error}</div>}
        <button type="submit" className="btn btn-primary">Login</button>
      </form>
      <div className="row">
        <button className="btn btn-secondary" onClick={async()=>{
          await fetch('/api/auth/resend',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({email})});
          setError('Verification email sent if account exists.');
        }}>Resend verification</button>
      </div>
    </div>
  );
}
