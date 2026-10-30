import { useEffect, useRef, useState } from "react";
import { ApiError } from "../../api/client";
import { useAuth } from "../../hooks/useAuth";

type CustomerVerificationModalProps = {
  onClose: () => void;
  onVerified: () => void;
};

function isValidIndianPhone(phone: string) {
  return /^[6-9]\d{9}$/.test(phone);
}

function getErrorMessage(error: unknown, fallback: string) {
  if (!(error instanceof ApiError)) {
    return fallback;
  }

  const message = error.message.toLowerCase();
  if (message.includes("expired")) {
    return "That code has expired. Request a new code and try again.";
  }
  if (message.includes("attempt") || message.includes("invalid") || error.status === 401) {
    return "That code is incorrect. Check it and try again.";
  }
  return error.message || fallback;
}

export function CustomerVerificationModal({ onClose, onVerified }: CustomerVerificationModalProps) {
  const { requestOtp, verifyOtp } = useAuth();
  const [step, setStep] = useState<"phone" | "otp">("phone");
  const [phone, setPhone] = useState("");
  const [otp, setOtp] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [resendAvailableAt, setResendAvailableAt] = useState<number | null>(null);
  const [now, setNow] = useState(() => Date.now());
  const phoneInputRef = useRef<HTMLInputElement>(null);
  const otpInputRef = useRef<HTMLInputElement>(null);

  const normalizedPhone = phone.replace(/\D/g, "").slice(-10);
  const canSend = isValidIndianPhone(normalizedPhone);
  const canVerify = /^\d{6}$/.test(otp);
  const resendSeconds = resendAvailableAt === null
    ? 0
    : Math.max(0, Math.ceil((resendAvailableAt - now) / 1000));

  useEffect(() => {
    phoneInputRef.current?.focus();
  }, []);

  useEffect(() => {
    if (step !== "otp" || resendAvailableAt === null) {
      return;
    }

    const timer = window.setInterval(() => {
      const currentTime = Date.now();
      setNow(currentTime);
      if (resendAvailableAt <= currentTime) {
        setResendAvailableAt(null);
      }
    }, 1000);

    return () => window.clearInterval(timer);
  }, [step, resendAvailableAt]);

  useEffect(() => {
    if (step === "otp") {
      otpInputRef.current?.focus();
    }
  }, [step]);

  async function handleRequestOtp(event: React.FormEvent) {
    event.preventDefault();
    if (!canSend || loading) {
      return;
    }

    setLoading(true);
    setError("");
    try {
      await requestOtp(`+91${normalizedPhone}`);
      setStep("otp");
      setResendAvailableAt(Date.now() + 30_000);
      setNow(Date.now());
    } catch (requestError) {
      setError(getErrorMessage(requestError, "We could not send the verification code. Please try again."));
    } finally {
      setLoading(false);
    }
  }

  async function handleVerifyOtp(event: React.FormEvent) {
    event.preventDefault();
    if (!canVerify || loading) {
      return;
    }

    setLoading(true);
    setError("");
    try {
      await verifyOtp(`+91${normalizedPhone}`, otp);
      onClose();
      onVerified();
    } catch (verificationError) {
      setError(getErrorMessage(verificationError, "We could not verify that code. Please try again."));
    } finally {
      setLoading(false);
    }
  }

  async function handleResend() {
    if (resendSeconds > 0 || loading) {
      return;
    }

    setLoading(true);
    setError("");
    try {
      await requestOtp(`+91${normalizedPhone}`);
      setOtp("");
      setResendAvailableAt(Date.now() + 30_000);
      setNow(Date.now());
    } catch (requestError) {
      setError(getErrorMessage(requestError, "We could not resend the code. Please try again."));
    } finally {
      setLoading(false);
    }
  }

  function changePhone() {
    setStep("phone");
    setOtp("");
    setError("");
    setResendAvailableAt(null);
  }

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center bg-slate-950/75 px-4 py-6 backdrop-blur-sm" role="dialog" aria-modal="true" aria-labelledby="customer-verification-title">
      <div className="w-full max-w-md rounded-[28px] border border-white/10 bg-white p-6 shadow-2xl sm:p-8">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-sm font-black uppercase tracking-[0.2em] text-orange-600">Secure checkout</p>
            <h2 id="customer-verification-title" className="mt-2 text-2xl font-black text-slate-950">
              {step === "phone" ? "Verify your mobile" : "Enter verification code"}
            </h2>
          </div>
          <button type="button" onClick={onClose} aria-label="Close verification" className="rounded-full px-2 text-2xl leading-none text-slate-400 transition hover:bg-slate-100 hover:text-slate-700">
            ×
          </button>
        </div>

        {step === "phone" ? (
          <form onSubmit={handleRequestOtp}>
            <p className="mt-4 text-sm leading-6 text-slate-600">Enter your Indian mobile number to continue with your cart.</p>
            <label htmlFor="customer-phone" className="mt-6 block text-sm font-bold text-slate-800">Mobile number</label>
            <div className="mt-2 flex rounded-2xl border border-slate-200 bg-slate-50 focus-within:ring-2 focus-within:ring-orange-500">
              <span className="flex items-center border-r border-slate-200 px-3 text-sm font-bold text-slate-600">+91</span>
              <input
                ref={phoneInputRef}
                id="customer-phone"
                value={normalizedPhone}
                onChange={(event) => setPhone(event.target.value.replace(/\D/g, "").slice(-10))}
                className="min-w-0 flex-1 bg-transparent px-3 py-3 outline-none"
                inputMode="numeric"
                autoComplete="tel-national"
                placeholder="9876543210"
                maxLength={10}
                aria-describedby={error ? "customer-auth-error" : undefined}
              />
            </div>
            {error ? <p id="customer-auth-error" role="alert" className="mt-3 rounded-2xl bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p> : null}
            <button type="submit" disabled={!canSend || loading} className="mt-6 w-full rounded-full bg-orange-600 px-5 py-3 text-sm font-black uppercase tracking-[0.15em] text-white transition hover:bg-orange-500 disabled:cursor-not-allowed disabled:opacity-50">
              {loading ? "Sending code…" : "Send OTP"}
            </button>
          </form>
        ) : (
          <form onSubmit={handleVerifyOtp}>
            <p className="mt-4 text-sm leading-6 text-slate-600">
              We sent a six-digit code to <span className="font-bold text-slate-800">+91 {normalizedPhone}</span>.
            </p>
            <label htmlFor="customer-otp" className="mt-6 block text-sm font-bold text-slate-800">Verification code</label>
            <input
              ref={otpInputRef}
              id="customer-otp"
              value={otp}
              onChange={(event) => setOtp(event.target.value.replace(/\D/g, "").slice(0, 6))}
              className="mt-2 w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-center text-2xl font-black tracking-[0.45em] outline-none focus:ring-2 focus:ring-orange-500"
              inputMode="numeric"
              autoComplete="one-time-code"
              maxLength={6}
              placeholder="••••••"
              aria-describedby={error ? "customer-auth-error" : undefined}
            />
            {error ? <p id="customer-auth-error" role="alert" className="mt-3 rounded-2xl bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p> : null}
            <button type="submit" disabled={!canVerify || loading} className="mt-6 w-full rounded-full bg-orange-600 px-5 py-3 text-sm font-black uppercase tracking-[0.15em] text-white transition hover:bg-orange-500 disabled:cursor-not-allowed disabled:opacity-50">
              {loading ? "Verifying…" : "Verify mobile number"}
            </button>
            <div className="mt-5 flex flex-wrap items-center justify-between gap-3 text-sm">
              <button type="button" onClick={changePhone} className="font-bold text-slate-600 underline underline-offset-4 hover:text-slate-900">Change number</button>
              <button type="button" onClick={() => void handleResend()} disabled={resendSeconds > 0 || loading} className="font-bold text-orange-700 disabled:cursor-not-allowed disabled:text-slate-400">
                {resendSeconds > 0 ? `Resend in ${resendSeconds}s` : "Resend code"}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
