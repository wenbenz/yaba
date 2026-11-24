import { Button } from '@mui/material';
import {DownloadOutlined} from "@ant-design/icons";

// Inside your RecentTransactionsTable component
export default function ExportButton() {
    const startDate = "0000-01-01";
    const endDate = (new Date()).toISOString().slice(0, 10);

    const handleExport = async () => {
        const params = new URLSearchParams({
            startDate: startDate,
            endDate: endDate
        });

        try {
            const response = await fetch(`/api/export/expenditure?${params}`);
            if (!response.ok) throw new Error('Export failed');

            // Get filename from response headers or use default
            const filename = response.headers.get('content-disposition')?.split('filename=')[1] || 'expenditure.csv';
            const blob = await response.blob();

            // Create download link and trigger it
            const url = window.URL.createObjectURL(blob);
            const link = document.createElement('a');
            link.href = url;
            link.download = filename;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            window.URL.revokeObjectURL(url);
        } catch (error) {
            console.error('Export failed:', error);
        }
    };

    return (
        <Button
            variant="contained"
            onClick={handleExport}
            startIcon={<DownloadOutlined />}
        >
            Export Expenditure
        </Button>
    );
};