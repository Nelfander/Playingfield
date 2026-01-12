type Props = {
    children: React.ReactNode
}

export default function Layout({ children }: Props) {
    return (
        <div
            style={{
                width: '100%',
                maxWidth: '420px',
                padding: '2rem',
            }}
        >
            {children}
        </div>
    )
}
