import React, { useEffect, useState } from 'react';
import { createClient, ConnectError } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { UserService } from '../../../gen/yaba/v1/user_pb';

const client = createClient(UserService, createConnectTransport({ baseUrl: '/rpc' }));

export default function Index() {
    const [email, setEmail] = useState('');
    const [savedEmail, setSavedEmail] = useState('');
    const [currentPassword, setCurrentPassword] = useState('');
    const [password, setPassword] = useState('');
    const [confirm, setConfirm] = useState('');
    const [status, setStatus] = useState(null);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        client.getProfile({}).then(({ profile }) => {
            setEmail(profile.email);
            setSavedEmail(profile.email);
        });
    }, []);

    const handleSubmit = async (e) => {
        e.preventDefault();
        setStatus(null);

        if (password && password !== confirm) {
            setStatus({ type: 'error', message: 'Passwords do not match' });
            return;
        }

        if (password && !currentPassword) {
            setStatus({ type: 'error', message: 'Current password is required' });
            return;
        }

        const emailChanged = email !== savedEmail;
        if (!emailChanged && !password) {
            setStatus({ type: 'error', message: 'Provide email and/or password to update' });
            return;
        }

        setLoading(true);
        try {
            if (emailChanged) {
                await client.updateProfile({ email });
                setSavedEmail(email);
            }
            if (password) await client.changePassword({ currentPassword, newPassword: password });
            setStatus({ type: 'success', message: 'Profile updated successfully' });
            setCurrentPassword('');
            setPassword('');
            setConfirm('');
        } catch (err) {
            setStatus({ type: 'error', message: ConnectError.from(err).rawMessage });
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
                        placeholder="email"
                        onChange={(e) => setEmail(e.target.value)}
                        style={{ width: '100%', padding: 8 }}
                    />
                </div>

                <div style={{ marginBottom: 12 }}>
                    <label style={{ display: 'block', marginBottom: 4 }}>Current Password</label>
                    <input
                        type="password"
                        value={currentPassword}
                        placeholder="required to change password"
                        onChange={(e) => setCurrentPassword(e.target.value)}
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
