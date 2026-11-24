import React, { useState } from 'react';

export default function Index() {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [confirm, setConfirm] = useState('');
    const [status, setStatus] = useState(null);
    const [loading, setLoading] = useState(false);

    const mutation = `
    mutation UpdateProfile($input: UpdateProfileInput!) {
      updateProfile(input: $input) {
        id
        username
        email
      }
    }
  `;

    const handleSubmit = async (e) => {
        e.preventDefault();
        setStatus(null);

        if (password && password !== confirm) {
            setStatus({ type: 'error', message: 'Passwords do not match' });
            return;
        }

        const input = {};
        if (email) input.email = email;
        if (password) input.password = password;

        if (!input.email && !input.password) {
            setStatus({ type: 'error', message: 'Provide email and/or password to update' });
            return;
        }

        setLoading(true);
        try {
            const res = await fetch('/graphql', {
                method: 'POST',
                credentials: 'include',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ query: mutation, variables: { input } }),
            });
            const json = await res.json();
            if (json.errors && json.errors.length) {
                setStatus({ type: 'error', message: json.errors[0].message || 'Update failed' });
            } else {
                setStatus({ type: 'success', message: 'Profile updated successfully' });
                setPassword('');
                setConfirm('');
            }
        } catch (err) {
            setStatus({ type: 'error', message: 'Network error' });
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{ maxWidth: 640, margin: '0 auto', padding: 16 }}>
            <h2>Profile Settings</h2>
            <form onSubmit={handleSubmit}>
                <div style={{ marginBottom: 12 }}>
                    <label style={{ display: 'block', marginBottom: 4 }}>Email</label>
                    <input
                        type="email"
                        value={email}
                        placeholder="new email (leave blank to keep current)"
                        onChange={(e) => setEmail(e.target.value)}
                        style={{ width: '100%', padding: 8 }}
                    />
                </div>

                <div style={{ marginBottom: 12 }}>
                    <label style={{ display: 'block', marginBottom: 4 }}>New Password</label>
                    <input
                        type="password"
                        value={password}
                        placeholder="new password (leave blank to keep current)"
                        onChange={(e) => setPassword(e.target.value)}
                        style={{ width: '100%', padding: 8 }}
                    />
                </div>

                <div style={{ marginBottom: 12 }}>
                    <label style={{ display: 'block', marginBottom: 4 }}>Confirm Password</label>
                    <input
                        type="password"
                        value={confirm}
                        placeholder="confirm new password"
                        onChange={(e) => setConfirm(e.target.value)}
                        style={{ width: '100%', padding: 8 }}
                    />
                </div>

                <button type="submit" disabled={loading} style={{ padding: '8px 16px' }}>
                    {loading ? 'Updating…' : 'Update'}
                </button>

                {status && (
                    <div style={{ marginTop: 12, color: status.type === 'error' ? 'crimson' : 'green' }}>
                        {status.message}
                    </div>
                )}
            </form>
        </div>
    );
}