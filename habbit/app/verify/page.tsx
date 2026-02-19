'use client';
import { useEffect, useState } from 'react';
import { useSearchParams } from 'next/navigation';

export default function Verify() {
  const params = useSearchParams();
  const token = params.get('token');
  const [status,setStatus] = useState('Verifying...');

  useEffect(()=>{
    if (!token) return;
    fetch(`/api/auth/verify?token=${token}`).then(async r=>{
      if (r.ok) setStatus('Verified. You can login.');
      else setStatus('Verification failed.');
    });
  },[token]);

  return <div className="card">{status}</div>;
}
